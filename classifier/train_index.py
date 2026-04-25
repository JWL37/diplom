import pandas as pd
import faiss
import json
import numpy as np
from sentence_transformers import SentenceTransformer
import os

CSV_FILE = 'goldens.csv'
INDEX_FILE = 'golden_index.faiss'
META_FILE = 'meta.json'

def train():
    print(f"Reading data from {CSV_FILE}...")
    df = pd.read_csv(CSV_FILE)
    
    # Отбрасываем пустые или битые строки
    df = df.dropna(subset=['text', 'label'])
    texts = df['text'].tolist()
    labels = df['label'].tolist()
    
    print("Loading model 'cointegrated/rubert-tiny2'...")
    model = SentenceTransformer('cointegrated/rubert-tiny2')
    
    print(f"Encoding {len(texts)} texts...")
    embeddings = model.encode(texts, show_progress_bar=True)
    
    # Нормализуем эмбеддинги для использования IndexFlatIP (косинусное сходство)
    # L2-нормализация каждого вектора (деление на его длину)
    embeddings = embeddings / np.linalg.norm(embeddings, axis=1)[:, None]
    
    print("Creating FAISS index...")
    dimension = embeddings.shape[1]
    # Используем Inner Product, так как вектора нормализованы, IP == Cosine Similarity
    index = faiss.IndexFlatIP(dimension)
    index.add(np.ascontiguousarray(embeddings))
    
    print("Saving index and metadata...")
    faiss.write_index(index, INDEX_FILE)
    
    meta = {
        str(i): {'text': texts[i], 'label': labels[i]}
        for i in range(len(texts))
    }
    with open(META_FILE, 'w', encoding='utf-8') as f:
        json.dump(meta, f, ensure_ascii=False, indent=2)
        
    print(f"Done! Index Size: {index.ntotal} vectors/goldens.")

if __name__ == '__main__':
    train()
