# Dealna AI Hybrid Search — Complete Technical Reference

> **Purpose:** Graduation project technical documentation for the AI-powered Hybrid Search Engine integrated into the Dealna Marketplace platform.

---

## 1. Problem Statement

University marketplaces serving Palestinian students require search that understands Arabic — including Modern Standard Arabic (MSA), Palestinian colloquial dialect, slang abbreviations, and mixed Arabic/English product names (e.g., "لابتوب Asus", "Calculus James Stewart", "مريول مختبر").

Traditional keyword search fails because:
- Arabic morphology is highly inflective (مريول ≠ مراويل, yet semantically identical)
- Students rarely spell product names the same way twice
- Typos and phonetic approximations are common in informal writing
- Dense-vector-only search fails on exact short keywords with low cosine similarity

**Solution:** A two-lane Hybrid Search architecture combining dense semantic vectors with fuzzy trigram matching, merged via Reciprocal Rank Fusion (RRF).

---

## 2. Full Infrastructure Stack

| Component | Technology | Region | Purpose |
|---|---|---|---|
| API Backend | Go (net/http + Chi) | EC2 `us-east-1` | REST API, business logic, orchestration |
| Relational DB | PostgreSQL 15 (pgx v5) | EC2 `us-east-1` | Source of truth for all item/user data |
| Embedding Worker | Python 3.11 (AWS Lambda) | `us-east-1` | AI model inference, vector upsert to Qdrant |
| Container Registry | AWS ECR | `us-east-1` | Hosts the Lambda Docker image |
| Vector Database | Qdrant Cloud | `us-east-1` | Dense vector storage, HNSW ANN search |
| Object Storage | AWS S3 | `us-east-1` | Item image storage (presigned URLs) |
| Identity Provider | Keycloak | EC2 `us-east-1` | JWT auth, university-scoped tenant tokens |
| Fuzzy Text Index | PostgreSQL `pg_trgm` | Same DB | Trigram GIN index for fuzzy keyword matching |

---

## 3. AI Model Selection

### Model: `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`

| Property | Value |
|---|---|
| Architecture | Transformer (MiniLM-L12) |
| Output Dimensions | **384** |
| Supported Languages | **50+ including Arabic** |
| Inference Engine | ONNX Runtime (via FastEmbed) |
| Model Size on Disk | ~120 MB |
| Inference Speed | ~15-30ms per query on Lambda |
| Distance Metric | Cosine Similarity |

**Why this model:**
- Native multilingual support means the model shares an embedding space across Arabic and English, enabling cross-lingual retrieval (searching "laptop" finds "لابتوب" items).
- ONNX runtime eliminates PyTorch overhead inside Lambda, keeping memory usage low enough to run in a 1024MB container.
- 384 dimensions is the optimal balance between accuracy and Qdrant RAM usage for a student-scale dataset.

---

## 4. Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      CLIENT (Mobile App)                         │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTP (Bearer JWT)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│               Go API Server (EC2 us-east-1 :8080)               │
│                                                                  │
│  ┌──────────────┐    ┌───────────────┐    ┌──────────────────┐  │
│  │ Auth         │    │ Item Handler  │    │ Search Handler   │  │
│  │ Middleware   │───▶│ POST /items   │    │ GET /items/search│  │
│  │ (Keycloak)  │    └───────┬───────┘    └────────┬─────────┘  │
│  └──────────────┘           │                     │             │
│                             │                     │             │
│  ┌──────────────────────────▼─────────────────────▼──────────┐  │
│  │                    ItemService                              │  │
│  │  CreateItem()                    SearchItems()             │  │
│  │      │                               │                     │  │
│  │      │ 1. Write to Postgres          │ concurrent goroutines│  │
│  │      │ 2. Fire Lambda (async)        ├──────┬──────────────│  │
│  │      │                              [A]    [B]             │  │
│  └──────┼───────────────────────────────┼──────┼─────────────┘  │
│         │                              │      │                  │
└─────────┼──────────────────────────────┼──────┼──────────────────┘
          │                              │      │
          │ [ASYNC Invoke]               │      │ [Postgres SQL]
          ▼                              ▼      ▼
 ┌────────────────┐     ┌────────────────────┐  ┌─────────────────┐
 │  AWS Lambda    │     │  AWS Lambda        │  │  PostgreSQL     │
 │  (Indexing)    │     │  (embed_query)     │  │  pg_trgm        │
 │                │     │  Synchronous       │  │  GIN Index      │
 │ • Embed text   │     │                    │  │                 │
 │ • Upsert to   │     │ • Embed query text │  │ similarity()    │
 │   Qdrant      │     │ • Return vector    │  │ ILIKE fallback  │
 └───────┬────────┘     └────────┬───────────┘  └────────┬────────┘
         │                       │                        │
         ▼                       ▼                        │
 ┌───────────────┐    ┌───────────────────────┐          │
 │  Qdrant Cloud │◀───│  Go: Query Qdrant     │          │
 │  (Vector DB)  │    │  gRPC + cosine ANN    │          │
 │  HNSW Index   │    │  score_threshold: 0.4 │          │
 │  384-dim      │    └─────────────┬─────────┘          │
 │  Cosine dist  │                  │                     │
 └───────────────┘                  │                     │
                                    ▼                     ▼
                         ┌──────────────────────────────────┐
                         │     RRF Merge in Go Backend      │
                         │   score = Σ 1/(60 + rank_i)     │
                         └─────────────────┬────────────────┘
                                           │
                                           ▼
                         ┌──────────────────────────────────┐
                         │  Postgres: GetFeedItemsByIDs()   │
                         │  WHERE id = ANY($1::uuid[])      │
                         │  ORDER BY array_position(...)    │
                         └─────────────────┬────────────────┘
                                           ▼
                              JSON Response to Client
```

---

## 5. Posting Flow (Item Creation → Vector Indexing)

### Step 1: API receives item creation request

```
POST /api/v1/items
Authorization: Bearer <keycloak_jwt>
```

### Step 2: Go writes to PostgreSQL first (Source of Truth)

```go
// item_service.go — CreateItem()
if err := s.repo.CreateItem(ctx, item); err != nil {
    return nil, err  // Never touch Lambda if DB write fails
}
// API returns 201 Created to the client BEFORE Lambda is invoked
```

### Step 3: SearchSyncEvent constructed and fired asynchronously

```go
sqsData := domain.SQSItemEventData{
    ItemID:      item.ID.String(),
    Title:       item.Title,        // Used ONLY for embedding generation
    Description: item.Description,  // Used ONLY for embedding generation
    Payload: domain.QdrantItemPayload{
        UniversityID: univID.String(), // Tenant partition key
        Category:     categoryName,
        Price:        item.Price,
        Status:       string(item.Status),
        Condition:    "used",
        IsGiveaway:   item.Price == 0,
    },
}
// Fire-and-forget — does not block the HTTP response
_ = s.searchSync.PublishSyncEvent(context.Background(), event)
```

### Step 4: Lambda invoked async (InvocationType: "Event")

```go
// lambda_publisher.go
go func() {
    p.client.Invoke(context.Background(), &lambda.InvokeInput{
        FunctionName:   &p.functionName,
        Payload:        payloadBytes,
        InvocationType: "Event", // AWS queues it, Go doesn't wait
    })
}()
```

### Step 5: Python Lambda generates embedding and upserts to Qdrant

```python
# lambda_function.py — action: "create"
text_to_embed = f"{title}. {description}"
vector = list(embedding_model.embed([text_to_embed]))[0].tolist()  # 384 floats

# Title and description are NOT stored — only the vector + metadata
qdrant_client.upsert(
    collection_name="dealna_items",
    points=[PointStruct(id=item_id, vector=vector, payload={
        "university_id": ..., "category": ..., "price": ...,
        "status": ..., "condition": ..., "is_giveaway": ...
    })]
)
```

> **Key Design Decision:** Text is **never persisted in Qdrant**. The 384-float vector encodes all semantic meaning. PostgreSQL is the single source of truth for text. This reduces Qdrant RAM usage by ~60%.

---

## 6. Search Flow (Hybrid: Dense + Keyword)

### Step 1: Request arrives

```
GET /api/v1/items/search?q=اجهزة كمبيوتر&min_price=0&max_price=500
Authorization: Bearer <keycloak_jwt>
```

### Step 2: University scoping (tenant isolation)

```go
univID, _ := s.repo.GetUniversityIDByUserID(ctx, filter.ExcludedOwnerID)
filter.RequesterUniversityID = univID
// All downstream queries scoped — users never see other universities
```

### Step 3: Two goroutines fire concurrently

```go
// Lane A: Dense semantic search (Lambda + Qdrant)
go func() {
    vector, _ := s.searchSync.GenerateEmbedding(ctx, query) // Sync Lambda call
    ids, _    := s.searchRepo.SearchItems(ctx, vector, filter) // Qdrant gRPC
    denseCh <- denseResult{ids: ids}
}()

// Lane B: Fuzzy keyword search (Postgres pg_trgm)
go func() {
    ids, _ := s.repo.KeywordSearchItems(ctx, query, filter)
    kwCh <- kwResult{ids: ids}
}()
```

### Step 4 (Lane A): Lambda vectorizes the query synchronously

```go
// InvocationType: "RequestResponse" — Go blocks, waits for the vector
res, _ := p.client.Invoke(ctx, &lambda.InvokeInput{
    InvocationType: "RequestResponse",
    Payload:        marshalledRequest, // {"action": "embed_query", "data": {"text": "اجهزة كمبيوتر"}}
})
// Response: {"statusCode": 200, "body": "{\"vector\": [0.021, -0.083, ...384 values...]}"}
```

```python
# lambda_function.py
vector = list(embedding_model.embed([query_text]))[0].tolist()
return {"statusCode": 200, "body": json.dumps({"vector": vector})}
```

### Step 5 (Lane A): Qdrant gRPC query

```go
// qdrant.go
var threshold float32 = 0.4  // Items below 0.4 cosine similarity rejected

q.Client.Query(ctx, &qdrant.QueryPoints{
    CollectionName: "dealna_items",
    Query:          qdrant.NewQuery(vector...),
    Filter:         &qdrant.Filter{Must: [university_id, status, price_range]},
    ScoreThreshold: &threshold,
    Limit:          &limit,
})
// Returns: UUIDs ranked by cosine similarity descending
```

### Step 6 (Lane B): Postgres pg_trgm query

```sql
SELECT i.item_id
FROM public.item i
JOIN public."User" u ON i.owner_id = u.user_id
WHERE u.university_id = $1
  AND i.item_status = 'AVAILABLE'
  AND i.deleted_at IS NULL
  AND i.owner_id != $2
  AND (
    similarity(i.title, $3) > 0.15
    OR similarity(i.description, $3) > 0.10
    OR i.title ILIKE '%' || $3 || '%'
    OR i.description ILIKE '%' || $3 || '%'
  )
ORDER BY GREATEST(similarity(i.title, $3), similarity(i.description, $3)) DESC
LIMIT $4
```

### Step 7: RRF Merge

```go
const rrfK = 60.0  // Standard Reciprocal Rank Fusion constant

scores := make(map[uuid.UUID]float64)

for rank, id := range denseRes.ids {
    scores[id] += 1.0 / (rrfK + float64(rank+1))
}
for rank, id := range kwRes.ids {
    scores[id] += 1.0 / (rrfK + float64(rank+1))
}
// Items appearing in BOTH lanes receive ~double score boost
// Sort descending → top items are definitively relevant
```

**Example — query "مريول":**
- Dense lane: 0 results (cosine below 0.4)
- Keyword lane: "مريول مختبر ونظارات حماية" → rank #1 via ILIKE
- RRF result: item returned correctly

**Example — query "اجهزة كمبيوتر":**
- Dense lane: Asus laptop, monitor, headphones (semantic)
- Keyword lane: items containing "كمبيوتر" in text
- RRF: overlapping items boosted to top

### Step 8: Single-query hydration from Postgres

```sql
-- No N+1 queries — one shot for all results
SELECT * FROM public.item i
WHERE i.item_id = ANY($1::uuid[])
ORDER BY array_position($1::uuid[], i.item_id)  -- Preserves RRF order
```

---

## 7. Lambda Deployment Architecture

### Why Docker Container?

Standard Lambda zip limit: **250MB**. Our dependencies total ~260MB:

| Dependency | Size |
|---|---|
| Python 3.11 runtime | ~50MB |
| ONNX Runtime | ~35MB |
| FastEmbed | ~25MB |
| MiniLM model weights | ~120MB |
| qdrant-client, boto3 | ~30MB |
| **Total** | **~260MB** |

Lambda Container Images support up to **10GB**.

### Dockerfile — Pre-baked Model Weights

```dockerfile
FROM python:3.11-slim

ENV FASTEMBED_CACHE_PATH=/app/fastembed_cache

RUN pip install --no-cache-dir awslambdaric
RUN pip install --no-cache-dir -r requirements.txt

# Download model at BUILD time — not at Lambda cold-start time
# Eliminates ~8s of model download on every cold start
RUN python -c "from fastembed import TextEmbedding; \
    TextEmbedding(model_name='sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2', \
    cache_dir='/app/fastembed_cache')"

COPY lambda_function.py .
ENTRYPOINT ["/usr/local/bin/python", "-m", "awslambdaric"]
CMD ["lambda_function.lambda_handler"]
```

### Lambda Configuration

| Setting | Value | Reason |
|---|---|---|
| Memory | 1024 MB | ONNX model requires ~400MB loaded |
| Timeout | 120 seconds | Batch indexing operations |
| Architecture | x86_64 | ONNX pre-compiled for x86 |
| Cold Start | ~10-12 seconds | Model loaded from pre-baked cache |
| Warm Start | ~150-300ms | Container stays alive after first invocation |

### Deployment Pipeline

```powershell
# --provenance=false required: Lambda rejects Docker BuildKit provenance manifests
docker build --provenance=false -t dealna-search-worker .
docker push 015615541352.dkr.ecr.us-east-1.amazonaws.com/dealna-search-worker:latest
aws lambda update-function-code --function-name dealna-search-worker --image-uri ...
```

---

## 8. Qdrant Collection Schema

**Collection:** `dealna_items` | **Distance:** Cosine | **Dimensions:** 384

| Payload Field | Index Type | Purpose |
|---|---|---|
| `university_id` | Keyword (`IsTenant=true`) | Physical HNSW graph partitioning per university |
| `status` | Keyword | Filter AVAILABLE/SOLD/RESERVED |
| `category` | Keyword | Optional category filter |
| `price` | Float | Range filter (min/max) |
| `condition` | Keyword | Item condition filter |
| `is_giveaway` | Bool | Filter free items |

> **`IsTenant=true` on `university_id`:** Physically partitions the HNSW ANN graph. Searches for Birzeit students never scan vectors from other universities — faster and cheaper at scale.

---

## 9. Status Update & Delete Flows

### Status Update

```python
# action: "update_status" — skips re-embedding, just updates payload
qdrant_client.set_payload(
    collection_name=COLLECTION_NAME,
    payload={"status": "SOLD"},
    points=[item_id]
)
```

### Delete

```python
# action: "delete" — hard delete from vector DB, soft delete in Postgres
qdrant_client.delete(
    collection_name=COLLECTION_NAME,
    points_selector=[item_id]
)
```

---

## 10. Trigram Extension Setup

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_item_title_trgm ON public.item USING GIN (title gin_trgm_ops);
CREATE INDEX idx_item_desc_trgm  ON public.item USING GIN (description gin_trgm_ops);
```

**How trigrams work:** "مريول" → trigrams `{"مري", "ريو", "يول"}`. The GIN index stores all trigrams from all items. Similarity = |shared trigrams| / |union of trigrams| (Jaccard). Works for Arabic because it operates at the character level — no tokenizer, no stemmer needed.

---

## 11. Security Design

| Concern | Solution |
|---|---|
| University isolation | `university_id` enforced in **both** Qdrant filter AND Postgres WHERE |
| Own item exclusion | `owner_id != requester_id` in Postgres SQL |
| API authentication | Keycloak JWT verified via JWKS endpoint on every request |
| AWS credentials | EC2 environment variables — no hardcoded keys |
| Qdrant access | Long-lived API key over TLS (gRPC port 6334) |

---

## 12. Performance Characteristics

| Scenario | Latency |
|---|---|
| Lambda cold start | ~10-12 seconds |
| Lambda warm embed | ~150-300ms |
| Qdrant gRPC query | ~5-15ms |
| Postgres trgm query | ~1-5ms |
| Full hybrid search (warm) | ~200-500ms |
| Full hybrid search (cold) | ~10-13 seconds |

---

## 13. API Reference

```
GET /api/v1/items/search
Authorization: Bearer <token>
```

| Parameter | Type | Description |
|---|---|---|
| `q` | string | Search query (Arabic/English) |
| `category_id` | UUID | Filter by category |
| `min_price` | float | Minimum price |
| `max_price` | float | Maximum price |
| `limit` | int | Results (default 20, max 100) |
| `offset` | int | Pagination offset |

Empty `q` → falls back to chronological feed.
Full Swagger: `http://localhost:8080/swagger/index.html`
