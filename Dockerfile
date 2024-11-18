# Use the official lightweight Python image
FROM python:3.13-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg git && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt ./
RUN python3 -m pip install --upgrade pip setuptools wheel && \
    pip3 install --no-cache-dir -r requirements.txt && \
    rm -rf ~/.cache/pip

COPY . .

CMD ["python3", "-m", "YukkiMusic"]
