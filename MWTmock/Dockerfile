FROM python:3.12.6-slim

WORKDIR /code

RUN pip install --upgrade pip

COPY ./requirements.txt requirements.txt

RUN pip install --no-cache-dir --upgrade --requirement requirements.txt

COPY ./app app

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
