FROM python:3.14-alpine

WORKDIR /usr/src/app

COPY . .

CMD ["python", "logs_reader.py"]
