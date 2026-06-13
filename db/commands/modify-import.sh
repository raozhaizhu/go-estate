sed -i '' 's/daily_data/`daily_data`/g' import-data.sql | head -n 20;
sed -i '' 's/"date"/`date`/g' import-data.sql | head -n 20;