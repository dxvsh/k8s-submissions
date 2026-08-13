import json
import time
import uuid
from datetime import datetime, timezone

LOG_FILE = "logs/logs.txt"
INTERVAL_SECONDS = 5
UUID_STRING = str(uuid.uuid4())


def current_timestamp():
    return datetime.now(timezone.utc).isoformat()


def write_log_line():
    log_entry = {
        "uuid": UUID_STRING,
        "timestamp": current_timestamp(),
    }

    with open(LOG_FILE, "a", encoding="utf-8") as log_file:
        log_file.write(json.dumps(log_entry) + "\n")
        log_file.flush()


def generate_logs():
    while True:
        write_log_line()
        time.sleep(INTERVAL_SECONDS)


if __name__ == "__main__":
    generate_logs()
