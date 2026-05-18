#!/usr/bin/env python3
import sqlite3
import pandas as pd # type: ignore
import os
import numpy
from typing import Tuple, Any
from sklearn.feature_extraction.text import TfidfVectorizer # type: ignore
from sklearn.preprocessing import StandardScaler # type: ignore
from sklearn.cluster import KMeans # type: ignore
from sklearn.decomposition import PCA # type: ignore
from sklearn.ensemble import RandomForestRegressor # type: ignore
from sklearn.model_selection import train_test_split # type: ignore
from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score # type: ignore
import numpy as np
import nltk # type: ignore
from nltk.corpus import stopwords # type: ignore
import matplotlib.pyplot as plt # type: ignore
import pickle # type: ignore

# Download Spanish stop words
try:
    nltk.data.find('corpora/stopwords')
except LookupError:
    nltk.download('stopwords')

# Disable scientific notation in pandas
pd.set_option('display.float_format', lambda x: f'{x:,.2f}')

# Get Spanish stop words
spanish_stopwords = set(stopwords.words('spanish'))


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
    df = df[df['price_amount'] >= limit]
    removed_count = initial_count - len(df)

    if removed_count > 0:
        print(f"Dropped {removed_count} products with price < {limit}")
        
    return df


def drop_higher_than(df: pd.DataFrame, limit: int) -> pd.DataFrame:
    # Drop products with price < limit
    initial_count = len(df)
    df = df[df['price_amount'] <= limit]
    removed_count = initial_count - len(df)

    if removed_count > 0:
        print(f"Dropped {removed_count} products with price > {limit}")
        
    return df


def currency_normalization(df: pd.DataFrame, limit: int, usd_price) -> pd.DataFrame:
    # Convert prices: if price > limit, divide by USD Price (currency normalization)
    df['price_amount'] = df['price_amount'].apply(
        lambda price: price / usd_price if price > limit else price
    )
    
    return df


def get_products(conn: sqlite3.Connection) -> pd.DataFrame:

    # Query only the three columns we need
    query = "SELECT title, description, price_amount FROM products"
    df = pd.read_sql_query(query, conn)

    # Ensure price_amount is float
    df['price_amount'] = pd.to_numeric(df['price_amount'], errors='coerce').astype('float64')
    
    df = currency_normalization(df, 20000, 1400)
    
    df = drop_lower_than(df, 200)
    
    df = drop_higher_than(df, 5000)

    return df


def get_highest_price(df: pd.DataFrame, count: int) -> pd.DataFrame:
    # Show top 5 highest priced products
    print(f"\n{'='*60}")
    print(f"Top {count} Highest Priced Products:")
    print(f"{'='*60}")
    top = df.nlargest(count, 'price_amount')[['title', 'price_amount']]
    for idx, (_, row) in enumerate(top.iterrows(), 1):
        print(f"{idx}. {row['title'][:70]} - ${row['price_amount']:,.2f}")
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
    print(df['price_amount'].describe())
    
    get_highest_price(df, 10)


def filter_price_outliers(df: pd.DataFrame) -> pd.DataFrame:
    # Remove price outliers (outside 2 std dev from mean)
    mean_price = df['price_amount'].mean()
    std_price = df['price_amount'].std()
    lower_bound = mean_price - (2 * std_price)
    upper_bound = mean_price + (2 * std_price)

    initial_count = len(df)
    df = df[(df['price_amount'] >= lower_bound) & (df['price_amount'] <= upper_bound)]
    removed_count = initial_count - len(df)

    print(f"\nPrice outlier removal:")
    print(f"  Mean: {mean_price:,.2f} | Std Dev: {std_price:,.2f}")
    print(f"  Valid range: {lower_bound:,.2f} - {upper_bound:,.2f}")
    print(f"  Removed {removed_count} outliers ({removed_count/initial_count*100:.1f}%)")
    print(f"  Remaining products: {len(df)}")
    
    return df


def classify_products(products_df: pd.DataFrame, categories_count: int = 5) -> None:
    """Classify products into N categories using title, description, and price."""

    print(f"\n{'='*60}")
    print(f"Classifying products into {categories_count} categories...")
    print(f"{'='*60}")

    # Copy dataframe
    df = products_df.copy()

    title_features, title_vectorizer = get_title_features(df)
    description_features, description_vectorizer = get_title_features(df)

    # Normalize price feature
    price_scaled = StandardScaler().fit_transform(df[['price_amount']])

    # Combine all features: title (50) + description (50) + price (1) = 101 dimensions
    features = np.hstack([title_features, description_features, price_scaled])
    print(f"Combined feature dimensions: {features.shape[1]}")

    # Apply K-means clustering
    print(f"Training K-means with {categories_count} clusters...")
    kmeans = KMeans(n_clusters=categories_count, random_state=42, n_init=10)
    df['category'] = kmeans.fit_predict(features)

    # Generate category names from model features
    title_feature_names = title_vectorizer.get_feature_names_out()
    description_feature_names = description_vectorizer.get_feature_names_out()
    category_names = {}

    print("\nGenerating category names from model features...")
    for category in range(categories_count):
        category_products = df[df['category'] == category]
        avg_price = category_products['price_amount'].mean()

        # Get the cluster center for this category
        center = kmeans.cluster_centers_[category]

        # Extract top words from title features (first 50)
        title_center = center[:len(title_feature_names)]
        top_title_idx = np.argsort(title_center)[-2:][::-1]
        top_title_words = [title_feature_names[i].title() for i in top_title_idx if i < len(title_feature_names)]

        # Extract top words from description features (next 50)
        desc_center = center[len(title_feature_names):len(title_feature_names)+len(description_feature_names)]
        top_desc_idx = np.argsort(desc_center)[-1:][::-1]
        top_desc_words = [description_feature_names[i].title() for i in top_desc_idx if i < len(description_feature_names)]

        # Create name from top words
        top_words = top_title_words + top_desc_words
        category_name = " & ".join(top_words[:3])

        # Add price tier
        if avg_price > 500000:
            category_name += " (Premium)"
        elif avg_price < 50000:
            category_name += " (Budget)"
        else:
            category_name += " (Mid-Range)"

        category_names[category] = category_name

    df['category_name'] = df['category'].map(category_names)

    # Plot the clusters
    plot_clusters(features, kmeans, df)

    # Show category distribution
    print(f"\n{'='*60}")
    print("Category Distribution:")
    print(f"{'='*60}")
    for category in range(categories_count):
        count = len(df[df['category'] == category])
        name = category_names[category]
        print(f"  {name}: {count} products")

    # Show products per category
    print(f"\n{'='*60}")
    print("Product Categories:")
    print(f"{'='*60}")

    for category in range(categories_count):
        category_products = df[df['category'] == category]
        avg_price = category_products['price_amount'].mean()
        count = len(category_products)
        name = category_names[category]

        print(f"\n📦 {name}")
        print(f"   Products: {count} | Avg Price: {avg_price:,.2f}")
        print(f"   Sample items:")

        # Show top 3 products
        for idx, (_, row) in enumerate(category_products.head(3).iterrows(), 1):
            print(f"      {idx}. {row['title'][:60]}")

    # Save classified data
    df.to_csv('products_classified.csv', index=False)
    print(f"\n✓ Classified data saved to products_classified.csv")
    

def plot_clusters(features: np.ndarray, kmeans: KMeans, df: pd.DataFrame) -> None:
    """Plot K-means clusters using PCA for 2D visualization."""

    print(f"\n{'='*60}")
    print("Plotting K-means clusters...")
    print(f"{'='*60}")

    # Reduce dimensions to 2D using PCA
    pca = PCA(n_components=2)
    features_2d = pca.fit_transform(features)

    # Create plot
    fig, ax = plt.subplots(figsize=(12, 8))

    # Plot each cluster with different color
    colors = plt.cm.Set3(np.linspace(0, 1, kmeans.n_clusters)) # type: ignore

    for cluster in range(kmeans.n_clusters):
        mask = df['category'] == cluster
        ax.scatter(
            features_2d[mask, 0],
            features_2d[mask, 1],
            c=[colors[cluster]],
            label=df[mask]['category_name'].iloc[0],
            alpha=0.6,
            s=50,
            edgecolors='black',
            linewidth=0.5
        )

    # Plot cluster centers
    centers_2d = pca.transform(kmeans.cluster_centers_)
    ax.scatter(
        centers_2d[:, 0],
        centers_2d[:, 1],
        c='red',
        marker='*',
        s=500,
        edgecolors='black',
        linewidth=2,
        label='Cluster Centers',
        zorder=5
    )

    ax.set_xlabel(f'PC1 ({pca.explained_variance_ratio_[0]:.1%} variance)')
    ax.set_ylabel(f'PC2 ({pca.explained_variance_ratio_[1]:.1%} variance)')
    ax.set_title(f'K-Means Clustering ({kmeans.n_clusters} clusters)', fontsize=14, fontweight='bold')
    ax.legend(loc='best', fontsize=10)
    ax.grid(alpha=0.3)

    plt.tight_layout()
    plt.savefig('clusters_visualization.png', dpi=150, bbox_inches='tight')
    print(f"✓ Cluster plot saved to clusters_visualization.png")
    plt.show(block=False)
    plt.pause(0.1)


def plot_prices(df: pd.DataFrame) -> None:
    """Plot price distribution with histograms and box plots."""

    print(f"\n{'='*60}")
    print("Plotting price distributions...")
    print(f"{'='*60}")

    fig, axes = plt.subplots(2, 2, figsize=(14, 10))
    fig.suptitle('Product Price Analysis', fontsize=16, fontweight='bold')

    # Plot 1: Histogram of all prices
    axes[0, 0].hist(df['price_amount'], bins=50, color='skyblue', edgecolor='black', alpha=0.7)
    axes[0, 0].set_xlabel('Price')
    axes[0, 0].set_ylabel('Frequency')
    axes[0, 0].set_title('Price Distribution (All Products)')
    axes[0, 0].grid(axis='y', alpha=0.3)

    # Plot 2: Box plot
    axes[0, 1].boxplot(df['price_amount'], vert=True)
    axes[0, 1].set_ylabel('Price')
    axes[0, 1].set_title('Price Box Plot')
    axes[0, 1].grid(axis='y', alpha=0.3)

    # Plot 3: Log scale histogram
    prices_nonzero = df[df['price_amount'] > 0]['price_amount']
    axes[1, 0].hist(prices_nonzero, bins=50, color='lightcoral', edgecolor='black', alpha=0.7)
    axes[1, 0].set_xlabel('Price (Log Scale)')
    axes[1, 0].set_ylabel('Frequency')
    axes[1, 0].set_yscale('log')
    axes[1, 0].set_xscale('log')
    axes[1, 0].set_title('Price Distribution (Log Scale)')
    axes[1, 0].grid(alpha=0.3)

    # Plot 4: Cumulative distribution
    sorted_prices = np.sort(df['price_amount'])
    cumulative = np.arange(1, len(sorted_prices) + 1) / len(sorted_prices)
    axes[1, 1].plot(sorted_prices, cumulative, linewidth=2, color='green')
    axes[1, 1].set_xlabel('Price')
    axes[1, 1].set_ylabel('Cumulative Probability')
    axes[1, 1].set_title('Cumulative Price Distribution')
    axes[1, 1].grid(alpha=0.3)

    plt.tight_layout()
    plt.savefig('price_distribution.png', dpi=150, bbox_inches='tight')
    print(f"✓ Plot saved to price_distribution.png")
    plt.show(block=False)
    plt.pause(0.1)


def get_title_features(df: pd.DataFrame) -> Tuple[numpy.ndarray, TfidfVectorizer]:
    # Vectorize title features
    print("Vectorizing title features...")
    title_vectorizer = TfidfVectorizer(
        max_features=50,
        stop_words=list(spanish_stopwords),
        lowercase=True
    )
    title_features = title_vectorizer.fit_transform(df['title'].fillna('')).toarray()
    return title_features, title_vectorizer

def get_description_features(df: pd.DataFrame) -> Tuple[numpy.ndarray, TfidfVectorizer]:
    # Vectorize description features
    print("Vectorizing description features...")
    description_vectorizer = TfidfVectorizer(
        max_features=50,
        stop_words=list(spanish_stopwords),
        lowercase=True
    )
    description_features = description_vectorizer.fit_transform(df['description'].fillna('')).toarray()
    return description_features, description_vectorizer


def save_model(model: Any, file_name: str) -> None:
    with open(file_name, 'wb') as f:
        pickle.dump(model, f)


def train_price_prediction_model(df: pd.DataFrame) -> tuple:
    """Train a regression model to predict product prices from title and description."""

    print(f"\n{'='*60}")
    print("Training Price Prediction Model...")
    print(f"{'='*60}")

    title_features, title_vectorizer = get_title_features(df)
    description_features, description_vectorizer = get_title_features(df)


    # Combine features
    features = np.hstack([title_features, description_features])
    target = df['price_amount'].values

    print(f"Combined feature dimensions: {features.shape[1]}")
    print(f"Training samples: {len(target)}")

    # Split data
    X_train, X_test, y_train, y_test = train_test_split(
        features, target, test_size=0.2, random_state=42
    )

    # Train model
    print("\nTraining Random Forest Regressor...")
    model = RandomForestRegressor(
        n_estimators=100,
        max_depth=20,
        min_samples_split=5,
        random_state=42,
        n_jobs=-1
    )
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
    for rank, idx in enumerate(top_features_idx, 1):
        if idx < len(title_vectorizer.get_feature_names_out()):
            feature_name = title_vectorizer.get_feature_names_out()[idx]
            source = "title"
        else:
            feature_name = description_vectorizer.get_feature_names_out()[idx - len(title_vectorizer.get_feature_names_out())]
            source = "description"
        importance = feature_importance[idx]
        print(f"  {rank}. {feature_name} ({source}): {importance:.4f}")


    save_model(model, 'price_prediction_model.pkl')
    save_model(title_vectorizer, 'title_vectorizer.pkl')
    save_model(description_vectorizer, 'description_vectorizer.pkl')
    
    print(f"\n✓ Model and vectorizers saved")

    return model, title_vectorizer, description_vectorizer, (X_test, y_test, y_pred_test)


def predict_product_prices(df: pd.DataFrame) -> pd.DataFrame:
    """Use trained model to predict prices for all products."""

    print(f"\n{'='*60}")
    print("Predicting Product Prices...")
    print(f"{'='*60}")

    # Load saved model and vectorizers
    try:
        with open('price_prediction_model.pkl', 'rb') as f:
            model = pickle.load(f)
        with open('title_vectorizer.pkl', 'rb') as f:
            title_vectorizer = pickle.load(f)
        with open('description_vectorizer.pkl', 'rb') as f:
            description_vectorizer = pickle.load(f)
        print("✓ Model and vectorizers loaded successfully")
    except FileNotFoundError as e:
        print(f"Error: Could not load model files. {e}")
        print("Please train the model first by running train_price_prediction_model()")
        return df

    # Copy dataframe to avoid modifying original
    result_df = df.copy()

    # Vectorize title using the saved vectorizer
    print("Vectorizing titles...")
    title_features = title_vectorizer.transform(result_df['title'].fillna('')).toarray()

    # Vectorize description using the saved vectorizer
    print("Vectorizing descriptions...")
    description_features = description_vectorizer.transform(result_df['description'].fillna('')).toarray()

    # Combine features
    features = np.hstack([title_features, description_features])
    print(f"Feature dimensions: {features.shape[1]}")

    # Make predictions
    print(f"Making predictions for {len(result_df)} products...")
    predicted_prices = model.predict(features)
    result_df['predicted_price'] = predicted_prices

    # Add prediction error (actual vs predicted)
    if 'price_amount' in result_df.columns:
        result_df['price_error'] = result_df['price_amount'] - result_df['predicted_price']
        result_df['price_error_pct'] = (result_df['price_error'] / result_df['price_amount'] * 100).round(2)

        # Show statistics
        print(f"\n{'='*60}")
        print("Prediction Statistics:")
        print(f"{'='*60}")
        print(f"Average predicted price: ${result_df['predicted_price'].mean():,.2f}")
        print(f"Average actual price: ${result_df['price_amount'].mean():,.2f}")
        mae = mean_absolute_error(result_df['price_amount'], result_df['predicted_price'])
        rmse = np.sqrt(mean_squared_error(result_df['price_amount'], result_df['predicted_price']))
        print(f"Mean Absolute Error: ${mae:,.2f}")
        print(f"RMSE: ${rmse:,.2f}")

        # Show biggest overestimates and underestimates
        print(f"\nTop 5 Overestimated (actual < predicted):")
        overest = result_df.nsmallest(5, 'price_error')[['title', 'price_amount', 'predicted_price', 'price_error_pct']]
        for idx, (_, row) in enumerate(overest.iterrows(), 1):
            print(f"  {idx}. {row['title'][:50]} | Actual: ${row['price_amount']:,.0f} | Predicted: ${row['predicted_price']:,.0f} ({row['price_error_pct']:.1f}%)")

        print(f"\nTop 5 Underestimated (actual > predicted):")
        underest = result_df.nlargest(5, 'price_error')[['title', 'price_amount', 'predicted_price', 'price_error_pct']]
        for idx, (_, row) in enumerate(underest.iterrows(), 1):
            print(f"  {idx}. {row['title'][:50]} | Actual: ${row['price_amount']:,.0f} | Predicted: ${row['predicted_price']:,.0f} ({row['price_error_pct']:.1f}%)")

    # Save predictions to CSV
    result_df.to_csv('products_with_predictions.csv', index=False)
    print(f"\n✓ Predictions saved to products_with_predictions.csv")

    return result_df


def main():
    conn = get_conn()
    products_df = get_products(conn)
    products_df = filter_price_outliers(products_df)

    df_statistics(products_df)
    plot_prices(products_df)
    classify_products(products_df, 7)
    train_price_prediction_model(products_df)
    predict_product_prices(products_df)

    conn.close()

    # Keep plot window open until user closes it
    print("\n✓ Analysis complete! Plot window will stay open. Close it to exit.")
    plt.show(block=True)


if __name__ == "__main__":
    main()