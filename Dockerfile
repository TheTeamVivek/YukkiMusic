FROM nikolaik/python-nodejs:python3.13-nodejs18

RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY . /app/
WORKDIR /app/
RUN pip3 install --no-cache-dir --upgrade pip setuptools \
 && pip3 install --no-cache-dir -U -r requirements.txt

CMD python3 -m YukkiMusic