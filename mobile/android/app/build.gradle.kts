import java.util.Properties

plugins {
    id("com.android.application")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

val keystorePropertiesFile = rootProject.file("key.properties")
val keystoreProperties: Properties? =
    if (keystorePropertiesFile.exists()) {
        Properties().apply { keystorePropertiesFile.inputStream().use { load(it) } }
    } else {
        null
    }

// A partial key.properties would otherwise fail later with an unclear error.
val keystoreKeys = listOf("keyAlias", "keyPassword", "storeFile", "storePassword")
keystoreProperties?.let { properties ->
    val missing = keystoreKeys.filter { properties.getProperty(it).isNullOrBlank() }
    if (missing.isNotEmpty()) {
        throw GradleException(
            "android/key.properties is missing ${missing.joinToString()} " +
                "(see mobile/README.md).",
        )
    }
}

android {
    namespace = "vn.letslive.mobile"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        applicationId = "vn.letslive.mobile"
        // Android 8.0 (API 26) is the product minimum.
        minSdk = 26
        targetSdk = flutter.targetSdkVersion
        // Uses the version code from pubspec.yaml. When using split APKs, 1000 * ABI_VERSION
        // is added automatically by Flutter. (https://developer.android.com/studio/build/configure-apk-splits#configure-APK-versions)
        // You can force using the value of versionCode by specifying the `-P force-version-code-ignoring-abi=true`
        // flag during build.
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    flavorDimensions += "env"
    productFlavors {
        create("dev") {
            dimension = "env"
            applicationIdSuffix = ".dev"
        }
        create("staging") {
            dimension = "env"
            applicationIdSuffix = ".staging"
        }
        create("prod") {
            dimension = "env"
        }
    }

    signingConfigs {
        if (keystoreProperties != null) {
            create("upload") {
                keyAlias = keystoreProperties.getProperty("keyAlias")
                keyPassword = keystoreProperties.getProperty("keyPassword")
                storeFile = rootProject.file(keystoreProperties.getProperty("storeFile"))
                storePassword = keystoreProperties.getProperty("storePassword")
            }
        }
    }

    buildTypes {
        release {
            // Release builds are signed with the keystore from key.properties
            // when it exists (required for staging and prod, enforced below);
            // without it the dev flavor falls back to the debug key.
            signingConfig = signingConfigs.findByName("upload")
                ?: signingConfigs.getByName("debug")
        }
    }
}

// Staging/prod release artifacts must never be signed with the debug key.
gradle.taskGraph.whenReady {
    val needsKey = allTasks.any {
        Regex("(assemble|bundle|package)(Staging|Prod)Release").containsMatchIn(it.name)
    }
    if (needsKey && keystoreProperties == null) {
        throw GradleException(
            "Release signing is not configured: android/key.properties is missing. " +
                "Create it with keyAlias, keyPassword, storeFile and storePassword " +
                "(see mobile/README.md). Staging and prod release builds are never " +
                "signed with the debug key.",
        )
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}
