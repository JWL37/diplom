from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import faiss
import json
import numpy as np
from sentence_transformers import SentenceTransformer
from collections import Counter
import os

app = FastAPI(title="Moderation Classifier API")

# Заглушки ресурсов для ленивой инициализации или при загрузке приложения
model = None
index = None
meta = None

INDEX_FILE = 'golden_index.faiss'
META_FILE = 'meta.json'


@app.on_event("startup")
def load_resources():
    global model, index, meta
    
    if not os.path.exists(INDEX_FILE) or not os.path.exists(META_FILE):
        print("WARNING: index or meta file not found. Ensure train_index.py was run.")
        return
        
    print("Loading SentenceTransformer model...")
    model = SentenceTransformer('cointegrated/rubert-tiny2')
    
    print(f"Loading FAISS index from {INDEX_FILE}...")
    index = faiss.read_index(INDEX_FILE)
    
    print(f"Loading metadata from {META_FILE}...")
    with open(META_FILE, 'r', encoding='utf-8') as f:
        meta = json.load(f)
    
    print("Resources loaded successfully.")


class ClassifyRequest(BaseModel):
    text: str
    k: int = 5

class ClassifyResponse(BaseModel):
    label: str
    confidence: float
    nearest_neighbors: list

@app.post("/classify", response_model=ClassifyResponse)
def classify_text(request: ClassifyRequest):
    if not model or not index or not meta:
        raise HTTPException(status_code=500, detail="Service not initialized properly (missing files).")
    
    text = request.text.strip()
    if not text:
        raise HTTPException(status_code=400, detail="Text cannot be empty.")
    
    # 1. Векторизуем текст
    embedding = model.encode([text])
    
    # 2. L2 нормализация для Inner Product (Косинусное сходство)
    embedding = embedding / np.linalg.norm(embedding, axis=1)[:, None]
    
    # 3. Поиск k ближайших соседей
    search_k = min(request.k, index.ntotal)
    distances, indices = index.search(np.ascontiguousarray(embedding), search_k)
    
    neighbors = []
    labels = []
    
    for d, i in zip(distances[0], indices[0]):
        # Отбрасываем отсутствующие индексы (-1)
        if i == -1:
            continue
            
        str_i = str(i)
        if str_i in meta:
            neighbor_meta = meta[str_i]
            lbl = neighbor_meta['label']
            labels.append(lbl)
            
            neighbors.append({
                "text": neighbor_meta['text'],
                "label": lbl,
                "distance": float(d)  # Cosine similarity in this case (higher is better)
            })
            
    if not labels:
        raise HTTPException(status_code=500, detail="No valid neighbors found.")
        
    # 4. Выбор мажоритарного класса (простое голосование)
    counter = Counter(labels)
    predicted_label, count = counter.most_common(1)[0]
    
    # Уверенность: доля соседей с победившим классом (от 0.0 до 1.0)
    confidence = count / len(labels)
    
    return ClassifyResponse(
        label=predicted_label,
        confidence=confidence,
        nearest_neighbors=neighbors
    )

if __name__ == '__main__':
    import uvicorn
    # Запуск отладочного сервера напрямую
    uvicorn.run(app, host="0.0.0.0", port=8000)
