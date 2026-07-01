SET
        NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS daily_data(
        `id` int4 PRIMARY KEY,
        `date` date NOT NULL,
        `region` smallint NOT NULL,
        `category` varchar(50) NOT NULL,
        `license_no` varchar(50) NOT NULL,
        `project_name` varchar(100) NOT NULL,
        `house_count` smallint NOT NULL,
        `area` decimal(10, 2) NOT NULL,
        `avg_price` decimal(10, 2),
        -- 日期查询 频率最高
        INDEX `idx_date` (`date`),
        -- 地区/分类/日期 复合查询
        INDEX `idx_region_category_date` (`region`, `category`, `date`),
        -- 预售证号精确查询
        INDEX `idx_license_no` (`license_no`),
        -- 备案名精确查询
        INDEX `idx_project_name` (`project_name`)
)