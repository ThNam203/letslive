FROM maven:3.9.9-eclipse-temurin-21-alpine AS build
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN mvn clean install -Dmaven.test.skip

FROM eclipse-temurin:17.0.13_11-jdk-alpine
WORKDIR /app
RUN apk --no-cache add curl
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8181
ENTRYPOINT ["java", "-jar", "app.jar"]
