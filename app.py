from flask import Flask
import logging

app = Flask(__name__)


@app.route("/")
def hello():
    return "Radhe Radhe"


log = logging.getLogger("werkzeug")
log.setLevel(logging.ERROR)
app.logger.disabled = True

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
