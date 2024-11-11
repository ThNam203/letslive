FROM eclipse-temurin:17.0.13_11-jdk-alpine
ARG JAR_FILE=target/*.jar
COPY ${JAR_FILE} app.jar
RUN apk --no-cache add curl
ENTRYPOINT ["java","-jar","/app.jar"]
