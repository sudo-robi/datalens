"""
DataLens Spark Analytics Job

Batch analytics pipeline that processes uploaded datasets and generates:
- Summary statistics (count, mean, min, max, nulls per column)
- Anomaly detection (z-score based outliers)
- Data quality metrics
"""

from pyspark.sql import SparkSession
from pyspark.sql import functions as F
from pyspark.sql.types import DoubleType
import json
import sys


def create_spark_session():
    return SparkSession.builder \
        .appName("DataLens Analytics") \
        .config("spark.jars.packages", "org.postgresql:postgresql:42.7.1") \
        .getOrCreate()


def compute_summary_statistics(df, numeric_columns):
    """Compute summary statistics for numeric columns."""
    stats = {}
    for col_name in numeric_columns:
        col_stats = df.select(
            F.count(col_name).alias("count"),
            F.countDistinct(col_name).alias("distinct_count"),
            F.mean(col_name).alias("mean"),
            F.stddev(col_name).alias("stddev"),
            F.min(col_name).alias("min"),
            F.max(col_name).alias("max"),
            F.expr(f"percentile_approx({col_name}, 0.5)").alias("median"),
            F.sum(F.when(F.col(col_name).isNull(), 1).otherwise(0)).alias("null_count"),
        ).collect()[0]

        stats[col_name] = {
            "count": col_stats["count"],
            "distinct_count": col_stats["distinct_count"],
            "mean": float(col_stats["mean"]) if col_stats["mean"] else None,
            "stddev": float(col_stats["stddev"]) if col_stats["stddev"] else None,
            "min": float(col_stats["min"]) if col_stats["min"] else None,
            "max": float(col_stats["max"]) if col_stats["max"] else None,
            "median": float(col_stats["median"]) if col_stats["median"] else None,
            "null_count": col_stats["null_count"],
        }

    return stats


def detect_anomalies(df, numeric_columns, threshold=3.0):
    """Detect anomalies using z-score method."""
    anomalies = {}

    for col_name in numeric_columns:
        stats = df.select(
            F.mean(col_name).alias("mean"),
            F.stddev(col_name).alias("stddev"),
        ).collect()[0]

        if stats["stddev"] is None or stats["stddev"] == 0:
            continue

        mean_val = stats["mean"]
        stddev_val = stats["stddev"]

        outliers = df.filter(
            F.abs((F.col(col_name) - mean_val) / stddev_val) > threshold
        ).count()

        anomalies[col_name] = {
            "outlier_count": outliers,
            "threshold": threshold,
            "mean": float(mean_val),
            "stddev": float(stddev_val),
        }

    return anomalies


def compute_data_quality(df):
    """Compute data quality metrics."""
    total_rows = df.count()
    total_cols = len(df.columns)

    column_quality = {}
    for col_name in df.columns:
        null_count = df.filter(F.col(col_name).isNull()).count()
        distinct_count = df.select(col_name).distinct().count()

        column_quality[col_name] = {
            "null_count": null_count,
            "null_percentage": (null_count / total_rows * 100) if total_rows > 0 else 0,
            "distinct_count": distinct_count,
            "completeness": ((total_rows - null_count) / total_rows * 100) if total_rows > 0 else 100,
        }

    return {
        "total_rows": total_rows,
        "total_columns": total_cols,
        "columns": column_quality,
    }


def run_analytics(spark, dataset_path, dataset_id):
    """Run full analytics pipeline on a dataset."""
    print(f"Processing dataset {dataset_id} from {dataset_path}")

    df = spark.read.csv(dataset_path, header=True, inferSchema=True)

    numeric_columns = [
        field.name for field in df.schema.fields
        if isinstance(field.dataType, DoubleType) or str(field.dataType) in ("IntegerType", "LongType", "FloatType", "DoubleType")
    ]

    results = {
        "dataset_id": dataset_id,
        "summary_statistics": compute_summary_statistics(df, numeric_columns),
        "anomalies": detect_anomalies(df, numeric_columns),
        "data_quality": compute_data_quality(df),
    }

    print(json.dumps(results, indent=2))
    return results


def main():
    if len(sys.argv) < 3:
        print("Usage: analytics_job.py <dataset_path> <dataset_id>")
        sys.exit(1)

    dataset_path = sys.argv[1]
    dataset_id = sys.argv[2]

    spark = create_spark_session()

    try:
        results = run_analytics(spark, dataset_path, dataset_id)
        print(f"Analytics complete for dataset {dataset_id}")
    finally:
        spark.stop()


if __name__ == "__main__":
    main()
