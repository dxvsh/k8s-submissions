import logging
import sys
import time
import uuid

logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
logger.propagate = False

handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(
    logging.Formatter(
        "%(asctime)s: %(message)s"
    )
)
logger.addHandler(handler)

UUID_STRING = str(uuid.uuid4())
INTERVAL = 5

def log_outputs():
    while True:
        logger.info(UUID_STRING)
        time.sleep(INTERVAL)


if __name__ == "__main__":
    log_outputs()
