#!/usr/bin/env python3
import sqlite3
import pandas as pd # type: ignore
import os

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


def main():
    conn = get_conn()
    
    df = get_products(conn)

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
    
    conn.close()

if __name__ == "__main__":
    main()