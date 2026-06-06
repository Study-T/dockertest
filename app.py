from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def root():
    return {"message": "Hello from dockertest!"}

@app.get("/health")
def health():
    return {"status": "ok"}