#!/usr/bin/env python3
import sqlite3
import pandas as pd # type: ignore
import os
from sklearn.feature_extraction.text import TfidfVectorizer # type: ignore
from sklearn.preprocessing import StandardScaler # type: ignore
from sklearn.cluster import KMeans # type: ignore
import numpy as np
import nltk # type: ignore
from nltk.corpus import stopwords # type: ignore

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


def get_products(conn: sqlite3.Connection) -> pd.DataFrame:

    # Query only the three columns we need
    query = "SELECT title, description, price_amount FROM products"
    df = pd.read_sql_query(query, conn)

    # Ensure price_amount is float
    df['price_amount'] = pd.to_numeric(df['price_amount'], errors='coerce').astype('float64')

    return df


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


def classify_products(products_df: pd.DataFrame, categories_count: int = 5) -> None:
    """Classify products into N categories using title, description, and price."""

    print(f"\n{'='*60}")
    print(f"Classifying products into {categories_count} categories...")
    print(f"{'='*60}")

    # Combine title and description
    df = products_df.copy()
    df['text'] = df['title'].fillna('') + ' ' + df['description'].fillna('')

    # Vectorize text using TF-IDF with Spanish stop words
    print("Vectorizing text features...")
    vectorizer = TfidfVectorizer(
        max_features=100,
        stop_words=list(spanish_stopwords),
        lowercase=True
    )
    text_features = vectorizer.fit_transform(df['text']).toarray()

    # Normalize price feature
    price_scaled = StandardScaler().fit_transform(df[['price_amount']])

    # Combine text + price features
    features = np.hstack([text_features, price_scaled])

    # Apply K-means clustering
    print(f"Training K-means with {categories_count} clusters...")
    kmeans = KMeans(n_clusters=categories_count, random_state=42, n_init=10)
    df['category'] = kmeans.fit_predict(features)

    # Generate category names from model features
    feature_names = vectorizer.get_feature_names_out()
    category_names = {}

    print("\nGenerating category names from model features...")
    for category in range(categories_count):
        category_products = df[df['category'] == category]
        avg_price = category_products['price_amount'].mean()

        # Get the cluster center for this category
        center = kmeans.cluster_centers_[category]

        # Extract top 3 words from TF-IDF features
        top_features_idx = np.argsort(center[:len(feature_names)])[-3:][::-1]
        top_words = [feature_names[i].title() for i in top_features_idx]

        # Create name from top words
        category_name = " & ".join(top_words)

        # Add price tier
        if avg_price > 500000:
            category_name += " (Premium)"
        elif avg_price < 50000:
            category_name += " (Budget)"
        else:
            category_name += " (Mid-Range)"

        category_names[category] = category_name

    df['category_name'] = df['category'].map(category_names)

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
    

def main():
    conn = get_conn()
    products_df = get_products(conn)
    
    df_statistics(products_df)
    classify_products(products_df)
    
    conn.close()


if __name__ == "__main__":
    main()