FROM python:3.13-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg git curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

ENV PATH="/root/.local/bin:${PATH}"

WORKDIR /app

COPY requirements.txt ./

RUN curl -LsSf https://astral.sh/uv/install.sh | sh && \
    uv self update && \
    uv pip install --upgrade --system setuptools wheel && \
    uv pip install --system -r requirements.txt

COPY . .

CMD ["python3", "-m", "YukkiMusic"]