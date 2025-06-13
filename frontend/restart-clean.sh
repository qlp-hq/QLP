#!/bin/bash

# Clean restart script for Next.js

echo "🧹 Cleaning Next.js cache..."
rm -rf .next
rm -rf node_modules/.cache

echo "📦 Ensuring dependencies are installed..."
npm install

echo "🚀 Starting development server..."
npm run dev
