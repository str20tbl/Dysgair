# Dysgair: Welsh Pronunciation Feedback Using Dual-Model ASR

MSc Language Technologies Dissertation by Stephen John Russell
Bangor University, 2024-2025

## Overview

This repository contains the implementation of a dual-model automatic speech recognition (ASR) system for providing character-level pronunciation feedback to Welsh L2 (second language) learners.

## Repository Structure

```
dysgair/
├── api/                                # Python FastAPI backend for ASR
│   ├── main.py                         # FastAPI endpoints for dual-model transcription
│   ├── statistical_analysis.py         # Statistical analysis module (3000+ LOC)
│   ├── requirements.txt                # Python dependencies (transformers, faster-whisper, etc.)
│   └── Dockerfile                      # API containerization
│
├── src/app/                            # Go Revel web application
│   ├── app/
│   │   ├── controllers/                # MVC controllers
│   │   │   ├── dysgair.go              # Recording upload/playback
│   │   │   ├── transcriptionReview.go  # Manual review interface
│   │   │   ├── analytics.go            # Statistical dashboard
│   │   │   └── userManagement.go       # User CRUD
│   │   ├── models/                     # Data models and algorithms
│   │   │   ├── attribution.go          # Error attribution logic
│   │   │   ├── metrics.go              # CER/WER calculations
│   │   │   └── levenshtein.go          # Character-level alignment
│   │   └── services/                   # Business logic layer
│   ├── public/                         # Frontend assets
│   │   ├── app/                        # JavaScript modules
│   │   │   ├── dysgair-recorder.js     # Web audio recording
│   │   │   ├── transcription-review.js # Admin annotation UI
│   │   │   └── analytics.js            # Chart visualization
│   │   └── images/                     # UI assets
│   ├── views/                          # HTML templates
│   ├── tests/                          # Test suite
│   ├── conf/                           # Revel configuration
│   └── go.mod                          # Go dependencies
│
├── scripts/                            # Utility scripts
├── docker-compose.yml                  # Full stack orchestration
└── README.md                           # This file
```

## Technology Stack

**Backend:**
- **Language:** Python 3.11+
- **Framework:** FastAPI (async REST API)
- **ML Models:**
  - OpenAI Whisper large-v3 (via faster-whisper)
  - Meta Wav2Vec2-XLS-R-300M (Hugging Face transformers)
- **Libraries:** NumPy, Pandas, SciPy, scikit-learn
- **Containerization:** Docker

**Frontend Web Application:**
- **Language:** Go 1.21+
- **Framework:** Revel (MVC web framework)
- **Frontend:** Vanilla JavaScript, HTML5, CSS3
- **Audio:** Web Audio API for browser-based recording
- **Charts:** Chart.js for analytics visualization

**Database:**
- **DBMS:** MySQL 8.0
- **ORM:** GORM (Go)
- **Migrations:** Revel db module

**Infrastructure:**
- **Containerization:** Docker Compose
- **Web Server:** Revel's built-in HTTP server
- **API Server:** Uvicorn (ASGI)

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)
- Python 3.11+ (for local API development)

### Running with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd dysgair

# Start all services (MySQL, API, Web App)
docker-compose up -d

# Web application will be available at:
# http://localhost:9000

# API documentation (Swagger UI) at:
# http://localhost:8000/docs
```
