Choose a backup folder in docker compose (volume), turn off the kong container (it uses the database container so the database  
wont start restore the backup)  

After that, run the docker compose file and exec the below command inside the database container

    pg_restore -U kong -C -d postgres --if-exists --clean kongdb_backup_20230816/
