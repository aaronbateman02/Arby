package matching

const embedWorkerScript = `#!/usr/bin/env python3
"""
embed_worker.py — standalone embedding worker for Arby

Fetches unembedded markets via the Arby REST API, embeds them locally
using BAAI/bge-large-en-v1.5, and writes embeddings back via the API.

Usage:
    pip install requests sentence-transformers torch
    python embed_worker.py --host https://arby.nostrabotus.com --batch-size 64
"""

from __future__ import annotations

import argparse
import logging
import sys
import time
from typing import Any

import requests
from sentence_transformers import SentenceTransformer

logging.basicConfig(
    format="%(asctime)s  %(levelname)-8s  %(message)s",
    datefmt="%H:%M:%S",
    level=logging.INFO,
)
log = logging.getLogger("embed_worker")

MODEL_NAME = "BAAI/bge-large-en-v1.5"


def load_model(batch_size: int) -> SentenceTransformer:
    log.info("Loading %s (batch_size=%d) ...", MODEL_NAME, batch_size)
    model = SentenceTransformer(MODEL_NAME)
    log.info("Model ready on %s", model.device)
    return model


def fetch_unembedded(host: str, limit: int) -> list[dict[str, Any]]:
    resp = requests.get(
        f"{host}/api/v1/markets/unembedded",
        params={"limit": limit},
        timeout=30,
    )
    resp.raise_for_status()
    data = resp.json()
    return data.get("markets", [])


def post_embeddings(host: str, embeddings: list[dict[str, Any]]) -> int:
    resp = requests.post(
        f"{host}/api/v1/markets/embeddings",
        json={"embeddings": embeddings},
        timeout=60,
    )
    resp.raise_for_status()
    return resp.json().get("updated", 0)


def main() -> None:
    parser = argparse.ArgumentParser(description="Arby embedding worker")
    parser.add_argument("--host", default="http://localhost:8087", help="Arby API base URL")
    parser.add_argument("--batch-size", type=int, default=64, help="Embedding batch size")
    parser.add_argument("--fetch-limit", type=int, default=2000, help="Markets per fetch")
    parser.add_argument("--sleep", type=int, default=10, help="Seconds to wait when queue is empty")
    args = parser.parse_args()

    host = args.host.rstrip("/")
    model = load_model(args.batch_size)

    total_embedded = 0
    empty_passes = 0

    log.info("Starting embed worker against %s", host)

    while True:
        try:
            batch = fetch_unembedded(host, args.fetch_limit)
        except requests.RequestException as e:
            log.warning("Fetch failed: %s - retrying in 30s", e)
            time.sleep(30)
            continue

        if not batch:
            empty_passes += 1
            log.info(
                "No unembedded markets (pass %d). Total: %d. Sleeping %ds ...",
                empty_passes, total_embedded, args.sleep,
            )
            time.sleep(args.sleep)
            continue

        empty_passes = 0
        log.info("Fetched %d unembedded markets", len(batch))

        texts = []
        for m in batch:
            title = (m.get("title") or "").strip()
            desc = (m.get("description") or "").strip()
            if desc:
                texts.append(f"{title}. {desc[:400]}")
            else:
                texts.append(title)

        log.info("Encoding %d texts ...", len(texts))
        vectors = model.encode(texts, batch_size=args.batch_size, normalize_embeddings=True, show_progress_bar=True)

        embeddings = [
            {"id": m["id"], "vector": v.tolist()}
            for m, v in zip(batch, vectors)
        ]

        try:
            updated = post_embeddings(host, embeddings)
        except requests.RequestException as e:
            log.warning("Upload failed: %s - retrying in 30s", e)
            time.sleep(30)
            continue

        total_embedded += updated
        log.info("Wrote %d embeddings (session total: %d)", updated, total_embedded)

        if updated < len(batch):
            log.warning("Expected %d updates, got %d - some may have failed", len(batch), updated)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        log.info("Interrupted")
        sys.exit(0)
`
