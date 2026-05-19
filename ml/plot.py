#!/usr/bin/env python3
import matplotlib
import matplotlib.pyplot as plt  # type: ignore
import numpy as np
import pandas as pd  # type: ignore
from sklearn.cluster import KMeans  # type: ignore
from sklearn.decomposition import PCA  # type: ignore

matplotlib.use("Agg")  # Use non-interactive backend to avoid tkinter cleanup errors


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
    colors = plt.cm.Set3(np.linspace(0, 1, kmeans.n_clusters))  # type: ignore

    for cluster in range(kmeans.n_clusters):
        mask = df["category_index"] == cluster
        ax.scatter(
            features_2d[mask, 0],
            features_2d[mask, 1],
            c=[colors[cluster]],
            label=df[mask]["category_name"].iloc[0],
            alpha=0.6,
            s=50,
            edgecolors="black",
            linewidth=0.5,
        )

    # Plot cluster centers
    centers_2d = pca.transform(kmeans.cluster_centers_)
    ax.scatter(
        centers_2d[:, 0],
        centers_2d[:, 1],
        c="red",
        marker="*",
        s=500,
        edgecolors="black",
        linewidth=2,
        label="Cluster Centers",
        zorder=5,
    )

    ax.set_xlabel(f"PC1 ({pca.explained_variance_ratio_[0]:.1%} variance)")
    ax.set_ylabel(f"PC2 ({pca.explained_variance_ratio_[1]:.1%} variance)")
    ax.set_title(
        f"K-Means Clustering ({kmeans.n_clusters} clusters)",
        fontsize=14,
        fontweight="bold",
    )
    ax.legend(loc="best", fontsize=10)
    ax.grid(alpha=0.3)

    plt.tight_layout()
    plt.savefig("clusters_visualization.png", dpi=150, bbox_inches="tight")
    print(f"✓ Cluster plot saved to clusters_visualization.png")
    plt.close()


def plot_prices(df: pd.DataFrame) -> None:
    """Plot price distribution with histograms and box plots."""

    print(f"\n{'='*60}")
    print("Plotting price distributions...")
    print(f"{'='*60}")

    fig, axes = plt.subplots(2, 2, figsize=(14, 10))
    fig.suptitle("Product Price Analysis", fontsize=16, fontweight="bold")

    # Plot 1: Histogram of all prices
    axes[0, 0].hist(
        df["price_usd"], bins=50, color="skyblue", edgecolor="black", alpha=0.7
    )
    axes[0, 0].set_xlabel("Price")
    axes[0, 0].set_ylabel("Frequency")
    axes[0, 0].set_title("Price Distribution (All Products)")
    axes[0, 0].grid(axis="y", alpha=0.3)

    # Plot 2: Box plot
    axes[0, 1].boxplot(df["price_usd"], vert=True)
    axes[0, 1].set_ylabel("Price")
    axes[0, 1].set_title("Price Box Plot")
    axes[0, 1].grid(axis="y", alpha=0.3)

    # Plot 3: Log scale histogram
    prices_nonzero = df[df["price_usd"] > 0]["price_usd"]
    axes[1, 0].hist(
        prices_nonzero, bins=50, color="lightcoral", edgecolor="black", alpha=0.7
    )
    axes[1, 0].set_xlabel("Price (Log Scale)")
    axes[1, 0].set_ylabel("Frequency")
    axes[1, 0].set_yscale("log")
    axes[1, 0].set_xscale("log")
    axes[1, 0].set_title("Price Distribution (Log Scale)")
    axes[1, 0].grid(alpha=0.3)

    # Plot 4: Cumulative distribution
    sorted_prices = np.sort(df["price_usd"])
    cumulative = np.arange(1, len(sorted_prices) + 1) / len(sorted_prices)
    axes[1, 1].plot(sorted_prices, cumulative, linewidth=2, color="green")
    axes[1, 1].set_xlabel("Price")
    axes[1, 1].set_ylabel("Cumulative Probability")
    axes[1, 1].set_title("Cumulative Price Distribution")
    axes[1, 1].grid(alpha=0.3)

    plt.tight_layout()
    plt.savefig("price_distribution.png", dpi=150, bbox_inches="tight")
    print(f"✓ Plot saved to price_distribution.png")
    plt.close()


def plot_description_length_distribution(df: pd.DataFrame) -> None:
    """Plot distribution of description lengths (character and word count)."""

    print(f"\n{'='*60}")
    print("Plotting Description Length Distribution...")
    print(f"{'='*60}")

    # Calculate description lengths
    df_copy = df.copy()
    df_copy["description_char_length"] = df_copy["description"].fillna("").str.len()
    df_copy["description_word_count"] = (
        df_copy["description"].fillna("").str.split().str.len()
    )

    # Print statistics
    print(f"\nDescription Character Length Statistics:")
    print(f"  Mean: {df_copy['description_char_length'].mean():.0f} characters")
    print(f"  Median: {df_copy['description_char_length'].median():.0f} characters")
    print(f"  Min: {df_copy['description_char_length'].min()} characters")
    print(f"  Max: {df_copy['description_char_length'].max()} characters")
    print(f"  Std Dev: {df_copy['description_char_length'].std():.0f}")

    print(f"\nDescription Word Count Statistics:")
    print(f"  Mean: {df_copy['description_word_count'].mean():.0f} words")
    print(f"  Median: {df_copy['description_word_count'].median():.0f} words")
    print(f"  Min: {df_copy['description_word_count'].min()} words")
    print(f"  Max: {df_copy['description_word_count'].max()} words")
    print(f"  Std Dev: {df_copy['description_word_count'].std():.0f}")

    # Create plot
    fig, axes = plt.subplots(2, 2, figsize=(14, 10))
    fig.suptitle("Product Description Length Analysis", fontsize=16, fontweight="bold")

    # Plot 1: Character length histogram
    axes[0, 0].hist(
        df_copy["description_char_length"],
        bins=50,
        color="skyblue",
        edgecolor="black",
        alpha=0.7,
    )
    axes[0, 0].set_xlabel("Character Length")
    axes[0, 0].set_ylabel("Frequency")
    axes[0, 0].set_title("Description Character Length Distribution")
    axes[0, 0].axvline(
        df_copy["description_char_length"].mean(),
        color="red",
        linestyle="--",
        linewidth=2,
        label="Mean",
    )
    axes[0, 0].axvline(
        df_copy["description_char_length"].median(),
        color="green",
        linestyle="--",
        linewidth=2,
        label="Median",
    )
    axes[0, 0].legend()
    axes[0, 0].grid(axis="y", alpha=0.3)

    # Plot 2: Word count histogram
    axes[0, 1].hist(
        df_copy["description_word_count"],
        bins=50,
        color="lightcoral",
        edgecolor="black",
        alpha=0.7,
    )
    axes[0, 1].set_xlabel("Word Count")
    axes[0, 1].set_ylabel("Frequency")
    axes[0, 1].set_title("Description Word Count Distribution")
    axes[0, 1].axvline(
        df_copy["description_word_count"].mean(),
        color="red",
        linestyle="--",
        linewidth=2,
        label="Mean",
    )
    axes[0, 1].axvline(
        df_copy["description_word_count"].median(),
        color="green",
        linestyle="--",
        linewidth=2,
        label="Median",
    )
    axes[0, 1].legend()
    axes[0, 1].grid(axis="y", alpha=0.3)

    # Plot 3: Box plot for character length
    axes[1, 0].boxplot(df_copy["description_char_length"], vert=True)
    axes[1, 0].set_ylabel("Character Length")
    axes[1, 0].set_title("Description Character Length Box Plot")
    axes[1, 0].grid(axis="y", alpha=0.3)

    # Plot 4: Box plot for word count
    axes[1, 1].boxplot(df_copy["description_word_count"], vert=True)
    axes[1, 1].set_ylabel("Word Count")
    axes[1, 1].set_title("Description Word Count Box Plot")
    axes[1, 1].grid(axis="y", alpha=0.3)

    plt.tight_layout()
    plt.savefig("description_length_distribution.png", dpi=150, bbox_inches="tight")
    print(f"\n✓ Plot saved to description_length_distribution.png")
    plt.close()


def plot_distance_distribution(df: pd.DataFrame) -> None:
    """Plot distribution of product distances from user location."""

    print(f"\n{'='*60}")
    print("Plotting Distance Distribution...")
    print(f"{'='*60}")

    # Filter out NaN distances
    df_with_distance = df[df["distance"].notna()].copy()

    if len(df_with_distance) == 0:
        print("No valid distance data to plot")
        return

    # Print statistics
    distances = df_with_distance["distance"]
    print(f"\nDistance Statistics (km):")
    print(f"  Mean: {distances.mean():.2f} km")
    print(f"  Median: {distances.median():.2f} km")
    print(f"  Std Dev: {distances.std():.2f} km")
    print(f"  Min: {distances.min():.2f} km")
    print(f"  Max: {distances.max():.2f} km")
    print(f"  Q1 (25%): {distances.quantile(0.25):.2f} km")
    print(f"  Q3 (75%): {distances.quantile(0.75):.2f} km")

    # Create plot
    fig, axes = plt.subplots(2, 2, figsize=(14, 10))
    fig.suptitle("Product Distance Distribution", fontsize=16, fontweight="bold")

    # Plot 1: Histogram of distances
    axes[0, 0].hist(distances, bins=50, color="steelblue", edgecolor="black", alpha=0.7)
    axes[0, 0].set_xlabel("Distance (km)")
    axes[0, 0].set_ylabel("Frequency")
    axes[0, 0].set_title("Distance Distribution")
    axes[0, 0].axvline(
        distances.mean(), color="red", linestyle="--", linewidth=2, label="Mean"
    )
    axes[0, 0].axvline(
        distances.median(), color="green", linestyle="--", linewidth=2, label="Median"
    )
    axes[0, 0].legend()
    axes[0, 0].grid(axis="y", alpha=0.3)

    # Plot 2: Box plot
    axes[0, 1].boxplot(distances, vert=True)
    axes[0, 1].set_ylabel("Distance (km)")
    axes[0, 1].set_title("Distance Box Plot")
    axes[0, 1].grid(axis="y", alpha=0.3)

    # Plot 3: Cumulative distribution
    sorted_distances = np.sort(distances)
    cumulative = np.arange(1, len(sorted_distances) + 1) / len(sorted_distances)
    axes[1, 0].plot(sorted_distances, cumulative, linewidth=2, color="purple")
    axes[1, 0].set_xlabel("Distance (km)")
    axes[1, 0].set_ylabel("Cumulative Probability")
    axes[1, 0].set_title("Cumulative Distance Distribution")
    axes[1, 0].grid(alpha=0.3)

    # Plot 4: Distance ranges (pie chart style)
    distance_ranges = pd.cut(
        distances,
        bins=[0, 5, 10, 25, 50, 100, float("inf")],
        labels=["0-5 km", "5-10 km", "10-25 km", "25-50 km", "50-100 km", ">100 km"],
    )
    range_counts = distance_ranges.value_counts().sort_index()

    colors = ["#2ecc71", "#3498db", "#f39c12", "#e74c3c", "#9b59b6", "#95a5a6"]
    axes[1, 1].bar(
        range(len(range_counts)),
        range_counts.values,
        color=colors[: len(range_counts)],
        edgecolor="black",
        alpha=0.7,
    )
    axes[1, 1].set_xticks(range(len(range_counts)))
    axes[1, 1].set_xticklabels(range_counts.index, rotation=45, ha="right")
    axes[1, 1].set_ylabel("Count")
    axes[1, 1].set_title("Products by Distance Range")
    axes[1, 1].grid(axis="y", alpha=0.3)

    plt.tight_layout()
    plt.savefig("distance_distribution.png", dpi=150, bbox_inches="tight")
    print(f"\n✓ Plot saved to distance_distribution.png")
    plt.close()
