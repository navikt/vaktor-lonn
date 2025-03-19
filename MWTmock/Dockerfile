FROM cgr.dev/chainguard/python:latest-dev as dev
WORKDIR /app
RUN pip install --upgrade pip
RUN python -m venv venv
ENV PATH="/app/venv/bin":$PATH
COPY ./requirements.txt /code/requirements.txt
RUN pip install --no-cache-dir --upgrade -r /code/requirements.txt

FROM cgr.dev/chainguard/python:latest
WORKDIR /app
COPY main.py main.py
COPY --from=dev /app /app
ENV PATH="/app/venv/bin:$PATH"
ENTRYPOINT ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
