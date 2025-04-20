FROM python:3.13-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg git && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt ./

RUN pip install --no-cache-dir uv && \
    uv self update && \
    uv pip install --upgrade setuptools wheel && \
    uv pip install -r requirements.txt

COPY . .

CMD ["python3", "-m", "YukkiMusic"]