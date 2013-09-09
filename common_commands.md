Teardown, recreate database
goschedule setup teardown --config config.json && goschedule setup create --config config.json

Run scrape, redirect stdout to stdin, pipe to tee (store in file and show as output)
goschedule scrape --config config.json 2>&1 | tee tmp.log