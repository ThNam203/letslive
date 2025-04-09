# ---- Stage 1: Build ----
FROM node:20-alpine AS builder

WORKDIR /home/node/app

# Copy package files first to leverage cache
COPY package*.json ./

# Use npm ci for faster, more reliable installs from package-lock.json
# Installs *all* dependencies (dev included) needed for the build
RUN npm ci

COPY . .

# Run the build script
RUN npm run build

# ---- Stage 2
# Start from a fresh, small base image
FROM node:18-alpine

WORKDIR /home/node/app

# Create app directory and change ownership (important for non-root execution)
# Do this *before* copying files as root, then switch user
RUN mkdir -p /home/node/app && chown -R node:node /home/node/app

# Copy only necessary package files for production install
COPY --chown=node:node package*.json ./

# Install *only* production dependencies.
# Use --omit=dev which is the modern equivalent of --only=production
RUN npm ci --omit=dev

# Copy built artifacts from the 'builder' stage
# Change '/home/node/app/dist' if your build output is in a different directory (e.g., build, public)
COPY --from=builder --chown=node:node /home/node/app/dist ./dist

# Copy other necessary runtime files like package.json (if needed for 'npm start')
# or specific configuration files that weren't part of the build output.
# COPY --from=builder --chown=node:node /home/node/app/package.json ./
# COPY --from=builder --chown=node:node /home/node/app/path/to/runtime/config.json ./config/

# Switch to the non-root user for security
USER node

# Expose the port your application listens on (replace 3000 if different)
EXPOSE 3000

CMD [ "node", "dist/index.js" ]
