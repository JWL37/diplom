import csv
import re
import os

INPUT_FILE = '../dataset.txt'
OUTPUT_FILE = 'goldens.csv'

def process_data():
    os.makedirs(os.path.dirname(OUTPUT_FILE) or '.', exist_ok=True)
    with open(INPUT_FILE, 'r', encoding='utf-8') as f_in, \
         open(OUTPUT_FILE, 'w', encoding='utf-8', newline='') as f_out:
        writer = csv.writer(f_out)
        writer.writerow(['text', 'label'])
        
        for line in f_in:
            line = line.strip()
            if not line:
                continue
            
            # Строка имеет формат: __label__INSULT,__label__THREAT текст комментария...
            # Разбиваем по первому пробелу: метки и сам текст
            parts = line.split(' ', 1)
            if len(parts) < 2:
                continue
            
            labels_str = parts[0]
            text = parts[1].strip()
            
            # Извлекаем первый попавшийся класс
            # пример: __label__INSULT,__label__THREAT -> INSULT
            match = re.search(r'__label__([A-Z]+)', labels_str)
            if match:
                label = match.group(1)
                writer.writerow([text, label])

if __name__ == '__main__':
    process_data()
    print(f"Data preparation complete. Saved to {OUTPUT_FILE}")
