FROM python:3.13-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg git curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt ./

RUN pip install -U uv && \
    uv pip install --upgrade --system setuptools wheel && \
    uv pip install --system -r requirements.txt

COPY . .

CMD ["python3", "-m", "YukkiMusic"]