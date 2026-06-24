#!/usr/bin/env python3
# Requires: pip install sentence-transformers requests

import argparse
import logging
import signal
import sys
import time

import requests
from sentence_transformers import SentenceTransformer

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

running = True


def handle_shutdown(signum, frame):
    global running
    logger.info("shutting down")
    running = False


def main():
    global running

    parser = argparse.ArgumentParser(description="Arby market embed worker")
    parser.add_argument(
        "--api-url",
        default="http://localhost:8086",
        help="Base URL of the Go monolith API",
    )
    parser.add_argument(
        "--model-name",
        default="BAAI/bge-large-en-v1.5",
        help="Sentence-transformers model name",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=64,
        help="Markets to fetch and embed per batch",
    )
    parser.add_argument(
        "--poll-interval",
        type=int,
        default=30,
        help="Seconds to sleep when no work",
    )
    parser.add_argument(
        "--rate-limit-sleep",
        type=int,
        default=60,
        help="Seconds to sleep on 429",
    )
    parser.add_argument(
        "--error-sleep",
        type=int,
        default=10,
        help="Seconds to sleep on error",
    )
    args = parser.parse_args()

    signal.signal(signal.SIGINT, handle_shutdown)
    signal.signal(signal.SIGTERM, handle_shutdown)

    logger.info("loading model %s", args.model_name)
    model = SentenceTransformer(args.model_name)
    model = model.to("cpu")
    logger.info("model loaded")

    while running:
        try:
            resp = requests.get(
                f"{args.api_url}/api/v1/markets/unembedded",
                params={"limit": args.batch_size},
                timeout=30,
            )

            if resp.status_code == 429:
                logger.warning(
                    "rate limited, sleeping %ds", args.rate_limit_sleep
                )
                time.sleep(args.rate_limit_sleep)
                continue

            resp.raise_for_status()
            data = resp.json()
            market_list = data.get("markets", [])
            if not market_list:
                logger.info("no unembedded markets")
                time.sleep(args.poll_interval)
                continue

            texts = []
            for market in market_list:
                embed_text = (
                    market["title"]
                    + ". "
                    + market.get("description", "")[:400]
                )
                texts.append(embed_text)

            embeddings = model.encode(
                texts, normalize_embeddings=True
            ).tolist()

            payload = {
                "embeddings": [
                    {"id": market["id"], "vector": emb}
                    for market, emb in zip(market_list, embeddings)
                ]
            }

            post_resp = requests.post(
                f"{args.api_url}/api/v1/markets/embeddings",
                json=payload,
                timeout=30,
            )

            if post_resp.status_code == 429:
                logger.warning(
                    "rate limited on post, sleeping %ds",
                    args.rate_limit_sleep,
                )
                time.sleep(args.rate_limit_sleep)
                continue

            post_resp.raise_for_status()
            logger.info("upserted %d embeddings", len(market_list))

        except Exception as e:
            logger.error("error: %s", e)
            if running:
                time.sleep(args.error_sleep)

    logger.info("shutting down")


if __name__ == "__main__":
    main()
