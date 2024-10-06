FROM python:3.12.7-slim

# Install necessary system dependencies
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       curl ffmpeg git build-essential libssl-dev apt-utils \
       zlib1g-dev libjpeg-dev libtiff5-dev libopenjp2-7 libtiff-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_19.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY . /app/
WORKDIR /app/
RUN python3 -m pip install --upgrade pip setuptools \
    && pip3 install --no-cache-dir --upgrade --requirement requirements.txt

CMD python3 -m YukkiMusic