FROM python:3.11-slim

# Install build tools just in case any package lacks a pre-compiled wheel
RUN apt-get update && apt-get install -y gcc g++ && rm -rf /var/lib/apt/lists/*

ENV FASTEMBED_CACHE_PATH=/app/fastembed_cache

WORKDIR /app
COPY requirements.txt .

# Install AWS Lambda Runtime Interface Client (RIC) and project dependencies
RUN pip install --no-cache-dir awslambdaric
RUN pip install --no-cache-dir -r requirements.txt

# Pre-download the multilingual model to eliminate cold start delays (Supports Arabic)
RUN python -c "from fastembed import TextEmbedding; TextEmbedding(model_name='sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2', cache_dir='/app/fastembed_cache')"

COPY lambda_function.py .

# Required Entrypoint for custom AWS Lambda images
ENTRYPOINT [ "/usr/local/bin/python", "-m", "awslambdaric" ]
CMD [ "lambda_function.lambda_handler" ]
