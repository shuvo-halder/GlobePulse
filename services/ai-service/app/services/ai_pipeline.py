import logging
from transformers import pipeline, AutoTokenizer, AutoModelForSequenceClassification
from sentence_transformers import SentenceTransformer
import numpy as np

logger = logging.getLogger(__name__)

class AIPipeline:
    def __init__(self):
        logger.info("Initializing AI Pipeline models...")
        # Summarization
        self.summarizer = pipeline("summarization", model="sshleifer/distilbart-cnn-12-6")
        
        # Sentiment Analysis
        self.sentiment_analyzer = pipeline("sentiment-analysis", model="distilbert-base-uncased-finetuned-sst-2-english")
        
        # NER for Entity and Country Extraction
        self.ner = pipeline("ner", aggregation_strategy="simple", model="dslim/bert-base-NER")
        
        # Zero-shot classification for Event Type and Topics
        self.classifier = pipeline("zero-shot-classification", model="facebook/bart-large-mnli")
        
        # Sentence embeddings for Similar News Detection and Clustering
        self.encoder = SentenceTransformer('all-MiniLM-L6-v2')
        logger.info("AI Pipeline models initialized.")

    def summarize(self, text: str) -> str:
        # Avoid breaking if text is too long or too short
        if len(text.split()) < 30:
            return text
        max_chunk = 1024
        text = text[:max_chunk]
        res = self.summarizer(text, max_length=130, min_length=30, do_sample=False)
        return res[0]['summary_text']

    def analyze_sentiment(self, text: str):
        max_chunk = 512
        text = text[:max_chunk]
        res = self.sentiment_analyzer(text)[0]
        score = res['score']
        label = res['label'] # POSITIVE or NEGATIVE
        # Map score to -1 to 1 roughly
        val = score if label == "POSITIVE" else -score
        return label, val

    def extract_entities_and_countries(self, text: str):
        res = self.ner(text[:1024])
        entities = []
        countries = set()
        for ent in res:
            entities.append({
                "name": ent['word'],
                "type": ent['entity_group'],
                "confidence": float(ent['score'])
            })
            if ent['entity_group'] == 'LOC':
                # Simplified country extraction, real implementation would map to ISO codes
                countries.add(ent['word'])
                
        return entities, list(countries)

    def detect_topics_and_events(self, text: str):
        topics_labels = ["politics", "economy", "technology", "health", "sports", "entertainment", "environment"]
        event_labels = ["election", "protest", "natural disaster", "conflict", "summit", "market crash", "policy announcement"]
        
        text_chunk = text[:1024]
        
        topics_res = self.classifier(text_chunk, topics_labels, multi_label=True)
        topics = [{"name": label, "score": score} for label, score in zip(topics_res['labels'], topics_res['scores']) if score > 0.5]
        
        event_res = self.classifier(text_chunk, event_labels)
        event_type = event_res['labels'][0] if event_res['scores'][0] > 0.4 else "general"
        
        return topics, event_type

    def generate_insights(self, text: str) -> str:
        # Simulated insight generation based on summarization and zero-shot
        return "AI analysis suggests significant regional impact based on current activity patterns."

    def score_importance(self, text: str) -> float:
        # Dummy heuristic for importance scoring based on length and keywords
        score = 0.5
        keywords = ["urgent", "breaking", "crisis", "war", "emergency", "global"]
        lower_text = text.lower()
        for kw in keywords:
            if kw in lower_text:
                score += 0.1
        return min(score, 1.0)
        
    def get_embedding(self, text: str) -> list[float]:
        embedding = self.encoder.encode(text)
        return embedding.tolist()

ai_pipeline = AIPipeline()

def get_ai_pipeline():
    return ai_pipeline
