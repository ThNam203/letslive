# ---- Stage 1: Build ----
# Use a specific Maven/JDK version, Alpine for smaller size
FROM maven:3.9.9-eclipse-temurin-21-alpine AS build
WORKDIR /app

# Copy POM file first to leverage dependency caching
COPY pom.xml .

# Download dependencies based on pom.xml before copying source
# This layer is cached unless pom.xml changes. Use -DskipTests for speed.
# Using 'verify' phase ensures dependencies needed for packaging are fetched.
RUN mvn dependency:go-offline -B -DskipTests=true
# Alternative: RUN mvn verify --fail-never -B -DskipTests=true # Also downloads plugins

# Copy source code
# Build is only re-triggered if source code changes (or pom.xml changes)
COPY src ./src

# Build the application JAR, skip tests. Use 'package'. '-B' for batch mode (no spinners)
RUN mvn clean package -B -DskipTests=true

# ---- Stage 2: Production ----
# Use a matching JRE version (Java 21 like build stage) on Alpine for smaller size
FROM eclipse-temurin:21-jre-alpine

WORKDIR /app

# Create a non-root user 'appuser' and group 'appgroup'
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# --- Optional: Install curl ONLY if absolutely needed by the runtime app ---
# If your application *executes* curl commands, uncomment the next line.
# Otherwise, remove it to keep the image small and secure.
RUN apk --no-cache add curl

# Copy the application JAR from the build stage, setting ownership to the new user
# Adjust the wildcard/name if your build produces a uniquely named JAR
COPY --from=build --chown=appuser:appgroup /app/target/*.jar app.jar

# Switch to the non-root user
USER appuser

# Expose the application port (ensure this matches your app's configuration)
EXPOSE 8181

# Set the entrypoint to run the JAR
# Consider adding JVM memory options (e.g., -Xmx, -Xms) if needed
ENTRYPOINT ["java", "-jar", "app.jar"]

# Optional: Add Healthcheck (requires Spring Boot Actuator or similar endpoint)
# Adjust the port and path (/actuator/health) as needed.
# HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
#   CMD ["java", "-jar", "healthcheck.jar"] # Or use curl if installed and preferred
# Example using curl (if installed):
# HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
#   CMD curl -f http://localhost:8181/actuator/health || exit 1
