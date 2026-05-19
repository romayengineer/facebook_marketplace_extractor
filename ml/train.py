#!/usr/bin/env python3
import os
import pickle  # type: ignore
import sqlite3
from dataclasses import dataclass, fields
from typing import Any, Dict, List, Optional, Tuple

import nltk  # type: ignore
import numpy
import numpy as np
import pandas as pd  # type: ignore
from dotenv import load_dotenv
from nltk.corpus import stopwords  # type: ignore
from sklearn.cluster import KMeans  # type: ignore
from sklearn.ensemble import RandomForestRegressor  # type: ignore
from sklearn.feature_extraction.text import TfidfVectorizer  # type: ignore
from sklearn.metrics import (  # type: ignore
    mean_absolute_error,
    mean_squared_error,
    r2_score,
)
from sklearn.model_selection import train_test_split  # type: ignore

# Download Spanish stop words
try:
    nltk.data.find("corpora/stopwords")
except LookupError:
    nltk.download("stopwords")

# Disable scientific notation in pandas
pd.set_option("display.float_format", lambda x: f"{x:,.2f}")

# Get Spanish stop words
spanish_stopwords = set(stopwords.words("spanish"))


@dataclass
class ModelsPKL:
    model: RandomForestRegressor
    title_vectorizer: TfidfVectorizer
    description_vectorizer: TfidfVectorizer
    category_vectorizer: Optional[TfidfVectorizer]


def get_conn() -> sqlite3.Connection:

    # Path to the PocketBase database
    db_path = "../pb_data/data.db"

    # Check if database exists
    if not os.path.exists(db_path):
        print(f"Error: Database not found at {db_path}")
        exit(1)

    # Connect to SQLite database
    conn = sqlite3.connect(db_path)

    return conn


def drop_lower_than(df: pd.DataFrame, limit: int) -> pd.DataFrame:
    # Drop products with price < limit
    initial_count = len(df)
    df = df[df["price_usd"] >= limit]
    removed_count = initial_count - len(df)

    if removed_count > 0:
        print(f"Dropped {removed_count} products with price < {limit}")

    return df


def drop_higher_than(df: pd.DataFrame, limit: int) -> pd.DataFrame:
    # Drop products with price < limit
    initial_count = len(df)
    df = df[df["price_usd"] <= limit]
    removed_count = initial_count - len(df)

    if removed_count > 0:
        print(f"Dropped {removed_count} products with price > {limit}")

    return df


def drop_description_len_higher_than(df: pd.DataFrame, limit: int) -> pd.DataFrame:
    # Drop products with description length > limit
    initial_count = len(df)
    df = df[df["description"].str.len() <= limit]
    removed_count = initial_count - len(df)

    if removed_count > 0:
        print(f"Dropped {removed_count} products with len(description) > {limit}")

    return df


def currency_normalization(df: pd.DataFrame, limit: int, usd_price) -> pd.DataFrame:
    # Convert prices: if price > limit, divide by USD Price (currency normalization)
    prices_above_limit_before = (df["price_amount"] > limit).sum()

    df["price_amount"] = df["price_amount"].map(lambda price: round(price, 2))

    df["price_usd"] = df["price_amount"].map(
        lambda price: round(price / usd_price if price > limit else price, 2)
    )

    # Report conversion results
    prices_above_limit_after = (df["price_usd"] > limit).sum()
    converted_count = prices_above_limit_before - prices_above_limit_after

    print(f"\nCurrency Normalization:")
    print(f"  Prices above {limit} before: {prices_above_limit_before}")
    print(f"  Prices above {limit} after:  {prices_above_limit_after}")
    print(f"  Successfully converted: {converted_count}")

    if converted_count <= 0:
        raise ValueError("did not convert")

    return df


def get_products(conn: sqlite3.Connection) -> pd.DataFrame:

    # Query only the three columns we need
    query = "SELECT id, title, description, category, price_amount, location_latitude, location_longitude FROM products"
    df = pd.read_sql_query(query, conn)

    clean_products(conn, df)

    # Ensure price_amount is float
    df["price_amount"] = pd.to_numeric(df["price_amount"], errors="coerce").astype(
        "float64"
    )

    df = currency_normalization(df, 4000, 1395)

    initial_count = len(df)
    print(f"Starting count of products {initial_count}")

    df = drop_lower_than(df, 50)

    df = drop_higher_than(df, 20000)

    df = drop_description_len_higher_than(df, 500)

    df = filter_price_outliers(df)

    df = calculate_distance(df)

    return df


def get_highest_price(df: pd.DataFrame, count: int) -> pd.DataFrame:
    # Show top 5 highest priced products
    print(f"\n{'='*60}")
    print(f"Top {count} Highest Priced Products:")
    print(f"{'='*60}")
    top = df.nlargest(count, "price_usd")[["title", "price_usd"]]
    for idx, (_, row) in enumerate(top.iterrows(), 1):
        print(f"{idx}. {row['title'][:70]} - ${row['price_usd']:,.2f}")
    return top


def df_statistics(df: pd.DataFrame) -> None:

    print(f"✓ Loaded {len(df)} products from database")
    print(f"\nDataset shape: {df.shape}")
    print(f"\nColumns: {df.columns.tolist()}")

    # Show first few rows
    print(f"\nFirst 5 rows:")
    print(df.head())

    # Show data info
    print(f"\nData info:")
    print(df.info())

    # Show missing values
    print(f"\nMissing values:")
    print(df.isnull().sum())

    # Show price statistics
    print(f"\nPrice statistics:")
    print(df["price_usd"].describe())

    get_highest_price(df, 10)


def filter_price_outliers(df: pd.DataFrame) -> pd.DataFrame:
    # Remove price outliers (outside 2 std dev from mean)
    mean_price = df["price_usd"].mean()
    std_price = df["price_usd"].std()
    lower_bound = mean_price - (2 * std_price)
    upper_bound = mean_price + (2 * std_price)

    initial_count = len(df)
    df = df[(df["price_usd"] >= lower_bound) & (df["price_usd"] <= upper_bound)]
    removed_count = initial_count - len(df)

    print(f"\nPrice outlier removal:")
    print(f"  Mean: {mean_price:,.2f} | Std Dev: {std_price:,.2f}")
    print(f"  Valid range: {lower_bound:,.2f} - {upper_bound:,.2f}")
    print(
        f"  Removed {removed_count} outliers ({removed_count/initial_count*100:.1f}%)"
    )
    print(f"  Remaining products: {len(df)}")

    return df


def calculate_distance(df: pd.DataFrame) -> pd.DataFrame:
    """Calculate distance from each product location to user location using Haversine formula."""

    print(f"\n{'='*60}")
    print("Calculating Distance from Products...")
    print(f"{'='*60}")

    # Load environment variables
    load_dotenv()
    my_latitude = float(os.environ["MY_LOCATION_LATITUDE"])
    my_longitude = float(os.environ["MY_LOCATION_LONGITUDE"])

    print(f"User location: ({my_latitude}, {my_longitude})")

    # Haversine formula to calculate distance between two lat/lon points
    def haversine(lat1: float, lon1: float, lat2: float, lon2: float) -> float:
        """Calculate distance in kilometers between two points."""
        from math import atan2, cos, radians, sin, sqrt

        R = 6371  # Earth radius in kilometers

        lat1_rad = radians(lat1)
        lat2_rad = radians(lat2)
        dlat = radians(lat2 - lat1)
        dlon = radians(lon2 - lon1)

        a = sin(dlat / 2) ** 2 + cos(lat1_rad) * cos(lat2_rad) * sin(dlon / 2) ** 2
        c = 2 * atan2(sqrt(a), sqrt(1 - a))

        return R * c

    # Calculate distance for each product
    df["distance"] = df.apply(
        lambda row: (
            haversine(
                my_latitude,
                my_longitude,
                row["location_latitude"],
                row["location_longitude"],
            )
            if pd.notna(row["location_latitude"])
            and pd.notna(row["location_longitude"])
            else np.nan
        ),
        axis=1,
    )

    # Print distance statistics
    valid_distances = df["distance"].dropna()
    print(f"\nDistance Statistics (km):")
    print(f"  Mean: {valid_distances.mean():.2f} km")
    print(f"  Median: {valid_distances.median():.2f} km")
    print(f"  Min: {valid_distances.min():.2f} km")
    print(f"  Max: {valid_distances.max():.2f} km")
    print(f"  Products with valid location: {len(valid_distances)}/{len(df)}")

    return df


def classify_products(products_df: pd.DataFrame, categories_count: int = 5) -> KMeans:
    """Classify products into N categories using title, description, and price."""

    print(f"\n{'='*60}")
    print(f"Classifying products into {categories_count} categories...")
    print(f"{'='*60}")

    df = products_df

    title_features, title_vectorizer = get_title_features(df)

    # Combine all features
    features = np.hstack([title_features])
    print(f"Combined feature dimensions: {features.shape[1]}")

    # Apply K-means clustering
    print(f"Training K-means with {categories_count} clusters...")
    kmeans = KMeans(n_clusters=categories_count, random_state=42, n_init=10)
    df["category_index"] = kmeans.fit_predict(features)

    # Generate category names from model features
    title_feature_names = title_vectorizer.get_feature_names_out()
    category_names = {}

    print("\nGenerating category names from model features...")
    for category in range(categories_count):
        category_products = df[df["category_index"] == category]
        avg_price = category_products["price_usd"].mean()

        # Get the cluster center for this category
        center = kmeans.cluster_centers_[category]

        # Extract top words from title features (first 50)
        title_center = center[: len(title_feature_names)]
        top_title_idx = np.argsort(title_center)[-2:][::-1]
        top_title_words = [
            title_feature_names[i].title()
            for i in top_title_idx
            if i < len(title_feature_names)
        ]

        # Create name from top words
        top_words = top_title_words
        category_name = "_".join(top_words[:5]).upper()

        category_names[category] = category_name

    df["category_name"] = df["category_index"].map(category_names)

    # Show category distribution
    print(f"\n{'='*60}")
    print("Category Distribution:")
    print(f"{'='*60}")
    for category in range(categories_count):
        count = len(df[df["category_index"] == category])
        name = category_names[category]
        print(f"  {name}: {count} products")

    # Show products per category
    print(f"\n{'='*60}")
    print("Product Categories:")
    print(f"{'='*60}")

    for category in range(categories_count):
        category_products = df[df["category_index"] == category]
        avg_price = category_products["price_usd"].mean()
        count = len(category_products)
        name = category_names[category]

        print(f"\n📦 {name}")
        print(f"   Products: {count} | Avg Price: {avg_price:,.2f}")
        print(f"   Sample items:")

        # Show top 3 products
        for idx, (_, row) in enumerate(category_products.head(3).iterrows(), 1):
            print(f"      {idx}. {row['title'][:60]}")

    # Save classified data
    df.to_csv("products_classified.csv", index=False)
    print(f"\n✓ Classified data saved to products_classified.csv")

    return kmeans


def get_title_features(df: pd.DataFrame) -> Tuple[numpy.ndarray, TfidfVectorizer]:
    # Vectorize title features
    print("Vectorizing title features...")
    title_vectorizer = TfidfVectorizer(
        max_features=100, stop_words=list(spanish_stopwords), lowercase=True
    )
    title_features = title_vectorizer.fit_transform(df["title"].fillna("")).toarray()
    return title_features, title_vectorizer


def get_description_features(df: pd.DataFrame) -> Tuple[numpy.ndarray, TfidfVectorizer]:
    # Vectorize description features
    print("Vectorizing description features...")
    description_vectorizer = TfidfVectorizer(
        max_features=50, stop_words=list(spanish_stopwords), lowercase=True
    )
    description_features = description_vectorizer.fit_transform(
        df["description"].fillna("")
    ).toarray()
    return description_features, description_vectorizer


def get_category_features(df: pd.DataFrame) -> Tuple[numpy.ndarray, TfidfVectorizer]:
    # Vectorize category features
    print("Vectorizing category features...")
    categories = df["category"].fillna("").astype(str).str.strip()

    # Handle empty categories
    if not categories.any() or (categories == "").all():
        print("  Warning: All categories are empty, using zeros")
        return np.zeros((len(df), 1)), None

    category_vectorizer = TfidfVectorizer(
        max_features=50, lowercase=True, min_df=1, stop_words=None
    )
    category_features = category_vectorizer.fit_transform(categories).toarray()
    return category_features, category_vectorizer


def save_model(model: Any, file_name: str) -> None:
    with open(file_name, "wb") as f:
        pickle.dump(model, f)


def save_pkls(category_name: str, models: ModelsPKL) -> None:
    for field in fields(models):
        value = getattr(models, field.name)
        if value == None:
            continue
        file_name = f"pkl/{category_name}_{field.name}.pkl"
        with open(file_name, "wb") as f:
            pickle.dump(value, f)


def load_pkls(category_name: str) -> ModelsPKL:
    models_data: Dict[str, Any] = {}
    for field in fields(ModelsPKL):
        file_name = f"pkl/{category_name}_{field.name}.pkl"
        if not os.path.exists(file_name):
            models_data[field.name] = None
            continue
        with open(file_name, "rb") as f:
            models = pickle.load(f)
            models_data[field.name] = models
    return ModelsPKL(**models_data)


def get_price_model() -> RandomForestRegressor:
    """
    n_estimators=100
    - Number of decision trees in the forest
    - More trees = more robust predictions but slower training/prediction
    - 100 is a reasonable default; more helps reduce overfitting but with diminishing returns

    max_depth=20
    - Maximum depth/height of each individual tree
    - Prevents trees from becoming too deep and overfitting to training data
    - Shallower trees (lower number) = simpler model, less prone to overfitting
    - Deeper trees (higher number) = captures more complex patterns but risk overfitting
    - 20 is moderate; allows trees to learn patterns without going too deep

    min_samples_split=5
    - Minimum number of samples required at a node to split it into two branches
    - Prevents the tree from creating splits on very small groups of data
    - Higher number = simpler, more generalizable model (less overfitting)
    - Lower number = tree can fit more complex patterns
    - 5 means "don't split a node unless it has at least 5 samples"

    random_state=42
    - Seed for the random number generator
    - Ensures reproducible results (same output every time you run it)
    - Without this, results vary slightly each run due to randomness in tree building
    - Any number works; 42 is just a convention (from Hitchhiker's Guide to the Galaxy!)

    n_jobs=-1
    - Number of CPU cores to use for parallel processing
    - -1 means use all available cores on your machine
    - Speeds up training significantly on multi-core systems
    - 1 would use a single core (slower but useful for debugging)
    """
    return RandomForestRegressor(
        n_estimators=2000, max_depth=40, min_samples_split=3, random_state=42, n_jobs=-1
    )


def train_price_prediction_model(
    df: pd.DataFrame, kmeans: KMeans
) -> Dict[str, ModelsPKL]:
    """Train a regression model to predict product prices from title and description."""

    df_for_category: Dict[str, ModelsPKL] = {}

    for cluster in range(kmeans.n_clusters):
        category_df = df[df["category_index"] == cluster]
        category_name: str = category_df["category_name"].iloc[0]

        print(f"\n{'='*60}")
        print(f"{category_name} Training Price Prediction Model...")
        print(f"{'='*60}")

        title_features, title_vectorizer = get_title_features(category_df)
        description_features, description_vectorizer = get_description_features(
            category_df
        )
        category_features, category_vectorizer = get_category_features(category_df)

        # Combine features: title (50) + description (50) + category (50) = 150 dimensions
        features = np.hstack([title_features, description_features, category_features])
        target = category_df["price_usd"].values

        print(f"Combined feature dimensions: {features.shape[1]}")
        print(f"Training samples: {len(target)}")

        # Split data
        X_train, X_test, y_train, y_test = train_test_split(
            features, target, test_size=0.2, random_state=42
        )

        # Train model
        print("\nTraining Random Forest Regressor...")
        model = get_price_model()
        model.fit(X_train, y_train)

        # Evaluate
        print("\nModel Evaluation:")
        y_pred_train = model.predict(X_train)
        y_pred_test = model.predict(X_test)

        train_rmse = np.sqrt(mean_squared_error(y_train, y_pred_train))
        test_rmse = np.sqrt(mean_squared_error(y_test, y_pred_test))
        train_mae = mean_absolute_error(y_train, y_pred_train)
        test_mae = mean_absolute_error(y_test, y_pred_test)
        train_r2 = r2_score(y_train, y_pred_train)
        test_r2 = r2_score(y_test, y_pred_test)

        print(f"  Train RMSE: {train_rmse:,.2f}")
        print(f"  Test RMSE:  {test_rmse:,.2f}")
        print(f"  Train MAE:  {train_mae:,.2f}")
        print(f"  Test MAE:   {test_mae:,.2f}")
        print(f"  Train R²:   {train_r2:.4f}")
        print(f"  Test R²:    {test_r2:.4f}")

        # Feature importance
        feature_importance = model.feature_importances_
        top_features_idx = np.argsort(feature_importance)[-10:][::-1]

        print(f"\nTop 10 Important Features:")
        title_len = len(title_vectorizer.get_feature_names_out())
        desc_len = len(description_vectorizer.get_feature_names_out())
        cat_len = (
            len(category_vectorizer.get_feature_names_out())
            if category_vectorizer
            else 0
        )

        for rank, idx in enumerate(top_features_idx, 1):
            if idx < title_len:
                feature_name = title_vectorizer.get_feature_names_out()[idx]
                source = "title"
            elif idx < title_len + desc_len:
                feature_name = description_vectorizer.get_feature_names_out()[
                    idx - title_len
                ]
                source = "description"
            elif category_vectorizer:
                feature_name = category_vectorizer.get_feature_names_out()[
                    idx - title_len - desc_len
                ]
                source = "category"
            else:
                feature_name = f"category_{idx - title_len - desc_len}"
                source = "category"
            importance = feature_importance[idx]
            print(f"  {rank}. {feature_name} ({source}): {importance:.4f}")

        models = ModelsPKL(
            model=model,
            title_vectorizer=title_vectorizer,
            description_vectorizer=description_vectorizer,
            category_vectorizer=category_vectorizer,
        )
        df_for_category[category_name] = models

        save_pkls(category_name, models)
        print(f"\n✓ Model and vectorizers saved")

    return df_for_category


def update_products_with_predictions(
    conn: sqlite3.Connection, result_df: pd.DataFrame
) -> None:
    """Update products table with prediction results (price_usd, predicted_price, price_error, price_error_pct)."""

    print(f"\n{'='*60}")
    print("Updating Products with Predictions...")
    print(f"{'='*60}")

    cursor = conn.cursor()

    update_data = result_df[
        [
            "id",
            "price_usd",
            "predicted_price",
            "price_error",
            "price_error_pct",
            "distance",
            "category_index",
            "category_name",
        ]
    ].copy()

    # Update products in database
    try:
        cursor.executemany(
            """UPDATE products
               SET price_usd = ?, predicted_price = ?, price_error = ?, price_error_pct = ?, distance = ?, category_index = ?, category_name = ?
               WHERE id = ?""",
            [
                (
                    row["price_usd"],
                    row["predicted_price"],
                    row["price_error"],
                    row["price_error_pct"],
                    row["distance"],
                    row["category_index"],
                    row["category_name"],
                    row["id"],
                )
                for _, row in update_data.iterrows()
            ],
        )

        conn.commit()
        print(f"✓ Updated {cursor.rowcount} products")

    except Exception as e:
        print(f"Error updating products: {e}")
        conn.rollback()

    finally:
        cursor.close()


def clean_products(conn: sqlite3.Connection, df: pd.DataFrame) -> None:

    print(f"\n{'='*60}")
    print("clean Products...")
    print(f"{'='*60}")

    cursor = conn.cursor()

    update_data = df[["id"]].copy()

    # Add columns with zeros for all rows
    columns_to_add = [
        "price_usd",
        "predicted_price",
        "price_error",
        "price_error_pct",
        "distance",
        "category_index",
        "category_name",
    ]
    for col in columns_to_add:
        update_data[col] = 0

    # Update products in database
    try:
        cursor.executemany(
            """UPDATE products
               SET price_usd = ?, predicted_price = ?, price_error = ?, price_error_pct = ?, distance = ?, category_index = ?, category_name = ?
               WHERE id = ?""",
            [
                (
                    row["price_usd"],
                    row["predicted_price"],
                    row["price_error"],
                    row["price_error_pct"],
                    row["distance"],
                    row["category_index"],
                    row["category_name"],
                    row["id"],
                )
                for _, row in update_data.iterrows()
            ],
        )

        conn.commit()
        print(f"✓ Updated {cursor.rowcount} products")

    except Exception as e:
        print(f"Error updating products: {e}")
        conn.rollback()

    finally:
        cursor.close()


def predict_product_prices(df: pd.DataFrame, kmeans: KMeans) -> pd.DataFrame:
    """Use trained model to predict prices for all products."""

    category_df_list: List[pd.DataFrame] = []

    for cluster in range(kmeans.n_clusters):
        category_df = df[df["category_index"] == cluster]
        category_name: str = category_df["category_name"].iloc[0]

        print(f"\n{'='*60}")
        print(f"{category_name} Predicting Product Prices...")
        print(f"{'='*60}")

        models = load_pkls(category_name)
        model = models.model
        title_vectorizer = models.title_vectorizer
        description_vectorizer = models.description_vectorizer
        category_vectorizer = models.category_vectorizer

        # Vectorize title using the saved vectorizer
        print("Vectorizing titles...")
        title_features = title_vectorizer.transform(
            category_df["title"].fillna("")
        ).toarray()

        # Vectorize description using the saved vectorizer
        print("Vectorizing descriptions...")
        description_features = description_vectorizer.transform(
            category_df["description"].fillna("")
        ).toarray()

        features_list = [title_features, description_features]

        if category_vectorizer:
            # Vectorize category using the saved vectorizer
            print("Vectorizing categories...")
            category_features = category_vectorizer.transform(
                category_df["category"].fillna("")
            ).toarray()
            features_list.append(category_features)
        else:
            # Add zeros for category features if no vectorizer (model was trained with them)
            category_features = np.zeros((len(category_df), 1))
            features_list.append(category_features)

        # Combine features: title (100) + description (50) + category (1 or more) = 151+ dimensions
        features = np.hstack(features_list)
        print(f"Feature dimensions: {features.shape[1]}")

        # Make predictions
        print(f"Making predictions for {len(category_df)} products...")
        predicted_prices = model.predict(features)
        category_df["predicted_price"] = predicted_prices.round(2)

        # Add prediction error (actual vs predicted)
        if "price_usd" in category_df.columns:
            category_df["price_error"] = (
                category_df["price_usd"] - category_df["predicted_price"]
            ).round(2)
            category_df["price_error_pct"] = (
                category_df["price_error"] / category_df["price_usd"] * 100
            ).round(2)

            # Show statistics
            print(f"\n{'='*60}")
            print("Prediction Statistics:")
            print(f"{'='*60}")
            print(
                f"Average predicted price: ${category_df['predicted_price'].mean():,.2f}"
            )
            print(f"Average actual price: ${category_df['price_usd'].mean():,.2f}")
            mae = mean_absolute_error(
                category_df["price_usd"], category_df["predicted_price"]
            )
            rmse = np.sqrt(
                mean_squared_error(
                    category_df["price_usd"], category_df["predicted_price"]
                )
            )
            print(f"Mean Absolute Error: ${mae:,.2f}")
            print(f"RMSE: ${rmse:,.2f}")

            # Show biggest overestimates and underestimates
            print(f"\nTop 5 Overestimated (actual < predicted):")
            overest = category_df.nsmallest(5, "price_error")[
                ["title", "price_usd", "predicted_price", "price_error_pct"]
            ]
            for idx, (_, row) in enumerate(overest.iterrows(), 1):
                print(
                    f"  {idx}. {row['title'][:50]} | Actual: ${row['price_usd']:,.0f} | Predicted: ${row['predicted_price']:,.0f} ({row['price_error_pct']:.1f}%)"
                )

            print(f"\nTop 5 Underestimated (actual > predicted):")
            underest = category_df.nlargest(5, "price_error")[
                ["title", "price_usd", "predicted_price", "price_error_pct"]
            ]
            for idx, (_, row) in enumerate(underest.iterrows(), 1):
                print(
                    f"  {idx}. {row['title'][:50]} | Actual: ${row['price_usd']:,.0f} | Predicted: ${row['predicted_price']:,.0f} ({row['price_error_pct']:.1f}%)"
                )

        category_df_list.append(category_df)

    result_df = pd.concat(category_df_list, ignore_index=True)

    # Save predictions to CSV
    result_df.to_csv("products_with_predictions.csv", index=False)
    print(f"\n✓ Predictions saved to products_with_predictions.csv")

    return result_df


def main():
    conn = get_conn()
    products_df = get_products(conn)

    # df_statistics(products_df)

    kmeans = classify_products(products_df, 5)
    train_price_prediction_model(products_df, kmeans)
    result_df = predict_product_prices(products_df, kmeans)
    update_products_with_predictions(conn, result_df)

    conn.close()


if __name__ == "__main__":
    main()
