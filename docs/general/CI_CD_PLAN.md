# CI/CD Pipeline for Arbitrum Sequencer Decoder using Woodpecker

## Overview
This document describes the Woodpecker CI/CD pipeline setup for the Arbitrum sequencer decoder project. The pipeline is designed to ensure code quality, security, and reliability throughout the development and deployment process.

## Woodpecker Configuration
The Woodpecker pipeline is configured in the `.woodpecker.yml` file in the repository root.

## Pipeline Structure

### 1. Development Pipeline
- Triggered on: pull requests to `main` and `develop` branches
- Purpose: Validate code changes before merge
- Components:
  - Format check
  - Linting
  - Unit tests
  - Security scan
  - Coverage analysis

### 2. Master/Release Pipeline
- Triggered on: pushes to `main` branch
- Purpose: Build, test, and deploy to production
- Components:
  - All development pipeline steps
  - Integration tests
  - Performance tests
  - Docker build and push
  - Deployment (production)

### 3. Release Pipeline
- Triggered on: Git tags matching semantic versioning
- Purpose: Create release artifacts and deploy
- Components:
  - All master pipeline steps
  - Binary build for multiple platforms
  - Release artifact creation
  - Docker build with version tags
  - GitHub release creation

## Pipeline Configuration

### .woodpecker.yml
This file will contain the actual Woodpecker pipeline configuration with the following stages:

- Format and linting stage
- Unit test stage
- Security scan stage
- Integration test stage
- Performance test stage
- Build stage
- Docker build and push stage
- Deploy stage

### Secrets Management
- Docker registry credentials
- Deployment target credentials
- Security scanning API keys
- Signing keys for release artifacts