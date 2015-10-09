-- TABLES
create table if not exists customer (
    id_customer serial primary key,
    name text,
    touched timestamp default clock_timestamp()
);

create table if not exists address (
    id_address serial primary key,
    description text,
    fk_customer int references customer
);

--
drop type if exists customer_type;
drop type if exists address_type;

create type address_type as (
    id_address int,
    description text
);

create type customer_type as (
    id_customer int,
    name text,
    addresses address_type[]
);

-- PROCEDURES
--- CUSTOMER_GET()
create or replace function customer_get(id int) returns json as $$
declare
    customer_json json;
begin
    select row_to_json(row(
            id_customer,
            name,
            (select coalesce(array_agg(row(id_address, description)::address_type), '{}') as address from address where fk_customer=id)
    )::customer_type)
    from customer
    where id_customer=id
    into customer_json;
    return customer_json;
end;
$$ language plpgsql;

--- CUSTOMER_WARMUP
create or replace function customer_warmup() returns boolean as $$
begin
    update customer set touched=clock_timestamp();
    return true;
end;
$$ language plpgsql;

-- TRIGGERS
--- CUSTOMER
create or replace function customer_updated() returns trigger as $$
begin
    if (tg_op = 'DELETE') then
        perform pg_notify('customer_deleted', old.id_customer::text);
    else
        perform pg_notify('customer_updated', new.id_customer::text);
    end if;
    return null;
end;
$$ language plpgsql;

drop trigger if exists customer_updated_trigger on customer;
create trigger customer_updated_trigger after insert or update or delete on customer for each row execute procedure customer_updated();

--- ADDRESS
create or replace function address_updated() returns trigger as $$
begin
    if (tg_op = 'DELETE') then
        perform pg_notify('customer_updated', old.fk_customer::text);
    else
        perform pg_notify('customer_updated', new.fk_customer::text);
    end if;
    return null;
end;
$$ language plpgsql;

drop trigger if exists address_updated_trigger on address;
create trigger address_updated_trigger after insert or update or delete on address for each row execute procedure address_updated();
