#!/usr/bin/env python3
import sqlite3
import pandas as pd # type: ignore
import os
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.preprocessing import StandardScaler
from sklearn.cluster import KMeans
import numpy as np

# Disable scientific notation in pandas
pd.set_option('display.float_format', lambda x: f'{x:,.2f}')


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

    # Vectorize text using TF-IDF
    print("Vectorizing text features...")
    vectorizer = TfidfVectorizer(max_features=100, stop_words='english', lowercase=True)
    text_features = vectorizer.fit_transform(df['text']).toarray()

    # Normalize price feature
    price_scaled = StandardScaler().fit_transform(df[['price_amount']])

    # Combine text + price features
    features = np.hstack([text_features, price_scaled])

    # Apply K-means clustering
    print(f"Training K-means with {categories_count} clusters...")
    kmeans = KMeans(n_clusters=categories_count, random_state=42, n_init=10)
    df['category'] = kmeans.fit_predict(features)

    # Show category distribution
    print(f"\nCategory Distribution:")
    print(df['category'].value_counts().sort_index())

    # Show sample products per category
    print(f"\n{'='*60}")
    print("Sample products per category:")
    print(f"{'='*60}")

    for category in range(categories_count):
        category_products = df[df['category'] == category]
        avg_price = category_products['price_amount'].mean()
        count = len(category_products)

        print(f"\n📦 Category {category} ({count} products, avg price: {avg_price:,.2f})")

        # Show top 3 most relevant products
        for idx, (_, row) in enumerate(category_products.head(3).iterrows()):
            print(f"   • {row['title'][:60]}")

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