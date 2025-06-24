# 🌐 DeepLink — Social Media Platform

**DeepLink** is a modern, scalable social media platform built for speed, reliability, and flexibility. Designed with a microservice architecture, it allows users to connect, share, and interact in real time while ensuring each system component remains independent, resilient, and easy to scale.

## 🚀 Key Features

- 🧑‍🤝‍🧑 Create and manage user profiles  
- 📝 Share text and multimedia posts  
- ❤️ Like, comment, and follow other users  
- 🔔 Real-time notifications  
- 💬 Messaging between users  
- 🌐 Distributed, event-driven backend


## 🧩 Architecture Overview

DeepLink is engineered with a **microservice architecture**, where each core functionality is handled by an independent service. These services communicate efficiently through a combination of synchronous and asynchronous technologies.

### 🔗 gRPC — Efficient Inter-Service Communication  
All internal services communicate using **gRPC**, a high-performance, language-neutral RPC framework. This enables low-latency communication with strong type safety using Protocol Buffers.

### 🐘 PostgreSQL — Structured Data Storage  
**PostgreSQL** is used to persist structured, relational data such as:
- Permissions and roles

### 📬 Kafka — Event-Driven Messaging  
**Apache Kafka** handles real-time, event-driven communication between services. Examples of Kafka events
- Post creation
- Notifications

This enables loose coupling between services and ensures that each one can process messages independently and at its own pace.

### ⚡ Redis — Caching and Session Management  
**Redis** significantly improves performance and user experience by reducing database load.

### 🍃 MongoDB — Unstructured Content Storage  
**MongoDB** is used for storing unstructured or semi-structured content such as:
- Images, videos, and media metadata
- User-generated messages
- Logs and histories

Its flexible schema makes it ideal for rapidly evolving content formats.

---

## 📚 Technologies Stack

| Area              | Technology     |
|-------------------|----------------|
| API Communication | gRPC           |
| Relational Data   | PostgreSQL     |
| Messaging Queue   | Apache Kafka   |
| Caching           | Redis          |
| NoSQL Storage     | MongoDB        |
| Architecture      | Microservices  |
| Containerization  | Docker + Docker Compose |



