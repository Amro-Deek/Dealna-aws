import os
import json
import logging
from typing import Any, Dict

from qdrant_client import QdrantClient
from qdrant_client.models import PointStruct
from fastembed import TextEmbedding

logger = logging.getLogger()
logger.setLevel(logging.INFO)

# 1. Initialize Clients OUTSIDE the handler for Lambda Container reuse (Warm Starts)
QDRANT_URL = os.environ.get("QDRANT_URL")
QDRANT_API_KEY = os.environ.get("QDRANT_API_KEY")
COLLECTION_NAME = "dealna_items"

if not QDRANT_URL or not QDRANT_API_KEY:
    raise ValueError("Missing QDRANT_URL or QDRANT_API_KEY environment variables.")

qdrant_client = QdrantClient(url=QDRANT_URL, api_key=QDRANT_API_KEY)

# Initialize the embedding model. FastEmbed automatically uses ONNX.
# We point it to the pre-downloaded cache baked into the Docker image.
# We are using paraphrase-multilingual-MiniLM-L12-v2 because it supports Arabic and is 384 dimensions (matching our Qdrant schema).
cache_dir = os.environ.get("FASTEMBED_CACHE_PATH", "/app/fastembed_cache")
embedding_model = TextEmbedding(model_name="sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2", cache_dir=cache_dir)


def _parse_event(event: Dict[str, Any]) -> list:
    """
    Normalize the incoming event into a list of search sync payloads.
    Supports two invocation modes:
      1. Direct Lambda Invoke (from Go goroutine): event IS the SearchSyncEvent itself
      2. SQS Trigger (legacy/fallback): event contains Records[] with body strings
    """
    if "Records" in event:
        # SQS format: each record's body is a JSON string
        payloads = []
        for record in event["Records"]:
            payloads.append(json.loads(record["body"]))
        return payloads
    else:
        # Direct invoke: the event itself is the payload
        return [event]


def lambda_handler(event: Dict[str, Any], context: Any) -> Dict[str, Any]:
    """
    AWS Lambda entry point. Accepts both direct invocation and SQS triggers.
    """
    payloads = _parse_event(event)
    logger.info(f"Processing {len(payloads)} search sync event(s).")

    for body in payloads:
        try:
            action = body.get("action")
            data = body.get("data", {})
            
            if action == "embed_query":
                query_text = data.get("text", "")
                if not query_text:
                    return {"statusCode": 400, "body": json.dumps({"error": "Missing 'text' for embed_query"})}
                
                logger.info(f"Generating embedding for query: '{query_text}'")
                embeddings_generator = embedding_model.embed([query_text])
                vector = list(embeddings_generator)[0].tolist()
                
                # Return immediately for synchronous invocation
                return {"statusCode": 200, "body": json.dumps({"vector": vector})}

            item_id = data.get("item_id")
            if not item_id:
                logger.error(f"Missing item_id in payload for action '{action}'.")
                continue

            logger.info(f"Processing action '{action}' for item '{item_id}'")

            if action == "delete":
                qdrant_client.delete(
                    collection_name=COLLECTION_NAME,
                    points_selector=[item_id]
                )
                logger.info(f"Deleted item {item_id} from Qdrant.")
                continue
            
            elif action in ["create", "update_status"]:
                # Combine Title and Description for rich semantic context
                title = data.get("title", "")
                description = data.get("description", "")
                
                if action == "create" and (not title and not description):
                    logger.warning(f"Item {item_id} has empty text fields. Skipping embedding.")
                    continue
                
                # Generate Embedding if this is a creation
                vector = None
                if action == "create":
                    text_to_embed = f"{title}. {description}"
                    # FastEmbed returns a generator of numpy arrays
                    embeddings_generator = embedding_model.embed([text_to_embed])
                    vector = list(embeddings_generator)[0].tolist()
                
                # Construct Payload (Strictly from the nested payload dictionary to match Go structs)
                payload_data = data.get("payload", {})
                
                payload = {
                    "university_id": payload_data.get("university_id"),
                    "category": payload_data.get("category"),
                    "price": float(payload_data.get("price", 0.0)),
                    "status": payload_data.get("status"),
                    "condition": payload_data.get("condition"),
                    "is_giveaway": payload_data.get("is_giveaway", False)
                }

                if vector:
                    # Full Upsert (New Item)
                    qdrant_client.upsert(
                        collection_name=COLLECTION_NAME,
                        points=[
                            PointStruct(
                                id=item_id,
                                vector=vector,
                                payload=payload
                            )
                        ]
                    )
                    logger.info(f"Upserted item {item_id} into Qdrant.")
                else:
                    # Partial Update (e.g., Status changed to 'sold' or 'reserved')
                    # We don't re-embed the text, just update the payload.
                    qdrant_client.set_payload(
                        collection_name=COLLECTION_NAME,
                        payload=payload,
                        points=[item_id]
                    )
                    logger.info(f"Updated payload for item {item_id} in Qdrant.")
            
            else:
                logger.warning(f"Unknown action: {action}")

        except Exception as e:
            logger.error(f"Error processing record: {str(e)}")
            raise e

    return {
        "statusCode": 200,
        "body": json.dumps(f"Successfully processed {len(payloads)} records.")
    }
