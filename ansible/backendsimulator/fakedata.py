import random
import os
import faker
import psycopg2

dsn = os.getenv('CP_DSN', 'postgresql://')
schemafile = os.getenv('CP_SCHEMA', 'schema.sql')

fake = faker.Faker()

db = psycopg2.connect(dsn)
c = db.cursor()

for batch in range(10):
    for customer in range(10):
        c.execute('insert into customer (name) values (%s) returning id_customer', (fake.name(),))
        id_customer = c.fetchone()
        for address in range(random.randint(1,4)):
            c.execute('insert into address (description, fk_customer) values (%s, %s)', (fake.address(), id_customer))
    db.commit()

c.close()
db.close()
