CREATE TABLE IF NOT EXISTS daily_data(
        id int4 PRIMARY KEY,
        date date NOT NULL,
        region smallint NOT NULL,
        category varchar(50) NOT NULL,
        license_no varchar(50) NOT NULL,
        project_name varchar(100) NOT NULL,
        house_count smallint NOT NULL,
        area decimal(10, 2) NOT NULL,
        avg_price decimal(10, 2))
