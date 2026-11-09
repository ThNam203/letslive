# Release shrink rules. Flutter's Gradle plugin applies this file to release
# builds automatically when it exists.

# WorkManager (pulled in by better_player_plus) builds its Room database by
# reflection on the generated WorkDatabase_Impl no-arg constructor. The Room
# 2.6 runtime it resolves to ships no keep rule for that constructor, so R8
# removes it and the app crashes on launch ("NoSuchMethodException:
# androidx.work.impl.WorkDatabase_Impl.<init> []"). Keep the constructor of
# every generated Room database.
-keep class * extends androidx.room.RoomDatabase {
    <init>();
}
