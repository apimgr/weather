# Use Node.js LTS (Long Term Support) version  
FROM node:20-alpine

# Set working directory
WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./

# Install dependencies
RUN npm ci --omit=dev && npm cache clean --force

# Create non-root user for security
RUN addgroup -g 1001 -S nodejs && \
    adduser -S weatherapp -u 1001

# Copy application code
COPY --chown=weatherapp:nodejs . .

# Create necessary directories
RUN mkdir -p /app/logs && \
    chown -R weatherapp:nodejs /app

# Switch to non-root user
USER weatherapp

# Expose port
EXPOSE 3000

# Health check using Kubernetes standard endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node -e "const http=require('http'); \
    http.get('http://localhost:3000/healthz', (res) => { \
      process.exit(res.statusCode === 200 ? 0 : 1); \
    }).on('error', () => process.exit(1));"

# Start the application
CMD ["npm", "start"]
