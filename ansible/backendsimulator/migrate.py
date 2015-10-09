#!/usr/bin/python
import os
import psycopg2

dsn = os.getenv('CP_DSN', 'postgresql://')
schemafile = os.getenv('CP_SCHEMA', 'schema.sql')

db = psycopg2.connect(dsn)
c = db.cursor()

c.execute(open(schemafile, 'r').read())
db.commit()

c.close()
db.close()
