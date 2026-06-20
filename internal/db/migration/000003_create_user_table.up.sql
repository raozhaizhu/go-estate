SET
        NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS users(
        id INT PRIMARY KEY AUTO_INCREMENT,
        -- 1. 业务字段
        username VARCHAR(50) UNIQUE NOT NULL,
        hashed_password VARCHAR(255) NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        -- 身份类型: 1.(注册用户,User, 默认); 2.(Vip, 会员); 3.( Admin, 管理员)
        role SMALLINT NOT NULL DEFAULT 1,
        -- 2. 时间字段
        -- 用于校验 token 是否发布于改变密码之前
        password_changed_at TIMESTAMP NOT NULL DEFAULT ('1970-01-01 00:00:01'),
        created_at TIMESTAMP NOT NULL DEFAULT (NOW())
)