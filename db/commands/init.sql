CREATE TABLE IF NOT EXISTS daily_data(
        id int4 PRIMARY KEY,
        date date,
        region smallint,
        category varchar(50),
        license_no varchar(50),
        project_name varchar(100),
        house_count smallint,
        area decimal(10, 2),
        avg_price decimal(10, 2))
