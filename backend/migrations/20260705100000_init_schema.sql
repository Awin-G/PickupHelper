-- +goose Up
-- +goose StatementBegin
CREATE DATABASE IF NOT EXISTS pickup_helper DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- +goose StatementEnd

-- 1. 驿站信息表
-- +goose StatementBegin
CREATE TABLE `stations` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '驿站ID',
  `name` varchar(100) NOT NULL COMMENT '驿站名称',
  `address` varchar(255) COMMENT '详细地址',
  `latitude` decimal(10,7) COMMENT '纬度',
  `longitude` decimal(10,7) COMMENT '经度',
  `business_hours` varchar(100) DEFAULT '09:00-20:00' COMMENT '营业时间',
  `status` tinyint DEFAULT 1 COMMENT '1-营业中, 0-休息中',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_name (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='驿站信息表';
-- +goose StatementEnd

-- 2. 用户表
-- +goose StatementBegin
CREATE TABLE `users` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
  `phone` varchar(11) NOT NULL COMMENT '手机号',
  `nickname` varchar(50) DEFAULT '' COMMENT '昵称',
  `avatar` varchar(255) DEFAULT '' COMMENT '头像URL',
  `openid` varchar(64) COMMENT '微信openid',
  `user_type` tinyint DEFAULT 1 COMMENT '1-普通收件人, 2-跑腿员',
  `runner_status` tinyint DEFAULT 0 COMMENT '0-未申请, 1-审核中, 2-已通过, 3-已拒绝',
  `credit_score` int DEFAULT 100 COMMENT '信用分',
  `is_blacklisted` tinyint DEFAULT 0 COMMENT '0-正常, 1-黑名单',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_phone (`phone`),
  UNIQUE KEY uk_openid (`openid`),
  INDEX idx_user_type (`user_type`),
  INDEX idx_runner_status (`runner_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
-- +goose StatementEnd

-- 3. 管理员表
-- +goose StatementBegin
CREATE TABLE `admins` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '管理员ID',
  `username` varchar(50) NOT NULL COMMENT '登录账号',
  `password_hash` varchar(255) NOT NULL COMMENT 'bcrypt加密密码',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `station_id` bigint COMMENT '所属驿站（系统管理员为NULL）',
  `real_name` varchar(50) COMMENT '真实姓名',
  `phone` varchar(11) COMMENT '联系电话',
  `status` tinyint DEFAULT 1 COMMENT '1-启用, 0-禁用',
  `last_login` datetime COMMENT '最后登录时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_username (`username`),
  INDEX idx_station (`station_id`),
  INDEX idx_status (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员表';
-- +goose StatementEnd

-- 4. 跑腿员申请表
-- +goose StatementBegin
CREATE TABLE `runner_applications` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '申请单ID',
  `user_id` bigint NOT NULL COMMENT '申请人用户ID',
  `real_name` varchar(50) NOT NULL COMMENT '真实姓名',
  `student_id` varchar(50) COMMENT '学号/工号',
  `id_card_image` varchar(255) COMMENT '证件照URL',
  `status` tinyint DEFAULT 1 COMMENT '1-审核中, 2-通过, 3-拒绝',
  `audit_admin_id` bigint COMMENT '审核管理员ID',
  `audit_remark` varchar(255) COMMENT '审核备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_user (`user_id`),
  INDEX idx_status (`status`),
  INDEX idx_created (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='跑腿员申请表';
-- +goose StatementEnd

-- 5. 包裹表（核心表）
-- +goose StatementBegin
CREATE TABLE `parcels` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '包裹ID',
  `station_id` bigint NOT NULL COMMENT '所属驿站',
  `tracking_no` varchar(64) NOT NULL COMMENT '快递单号',
  `courier_company` varchar(50) NOT NULL COMMENT '快递公司',
  `shelf_code` varchar(20) COMMENT '货架编号',
  `pickup_code` varchar(10) NOT NULL COMMENT '6位取件码',
  `receiver_phone` varchar(11) NOT NULL COMMENT '收件人手机号',
  `receiver_user_id` bigint COMMENT '关联注册用户ID',
  `receiver_name` varchar(50) COMMENT '收件人姓名',
  `weight` decimal(10,2) DEFAULT 0 COMMENT '重量(kg)',
  `is_fragile` tinyint DEFAULT 0 COMMENT '0-普通, 1-易碎',
  `remarks` varchar(255) COMMENT '备注',
  `status` tinyint DEFAULT 1 COMMENT '1-待取, 2-已取, 3-滞留, 4-已退件, 5-异常',
  `storage_time` datetime NOT NULL COMMENT '入库时间',
  `pickup_time` datetime COMMENT '取件时间',
  `return_time` datetime COMMENT '退件时间',
  `last_notify_time` datetime COMMENT '最近一次催取时间',
  `notify_count` int DEFAULT 0 COMMENT '催取次数',
  `operator_id` bigint COMMENT '入库管理员ID',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tracking_station (`tracking_no`, `station_id`),
  UNIQUE KEY uk_pickup_code_station (`pickup_code`, `station_id`),
  INDEX idx_receiver_phone (`receiver_phone`),
  INDEX idx_receiver_user_id (`receiver_user_id`),
  INDEX idx_status (`status`),
  INDEX idx_shelf (`shelf_code`),
  INDEX idx_storage_time (`storage_time`),
  INDEX idx_station (`station_id`),
  INDEX idx_station_status (`station_id`, `status`),
  INDEX idx_station_storage_time (`station_id`, `storage_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='包裹表';
-- +goose StatementEnd

-- 6. 取件日志表
-- +goose StatementBegin
CREATE TABLE `pickup_logs` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
  `parcel_id` bigint NOT NULL COMMENT '包裹ID',
  `operator_id` bigint COMMENT '操作人ID',
  `operator_type` tinyint NOT NULL COMMENT '1-管理员, 2-自助机, 3-跑腿员, 4-本人',
  `verification_method` tinyint NOT NULL COMMENT '1-扫码, 2-手动输入, 3-人脸',
  `location_lat` decimal(10,7) COMMENT '取件位置纬度',
  `location_lng` decimal(10,7) COMMENT '取件位置经度',
  `ip_address` varchar(45) COMMENT 'IP地址',
  `user_agent` varchar(255) COMMENT '客户端标识',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_parcel (`parcel_id`),
  INDEX idx_operator (`operator_id`),
  INDEX idx_created (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='取件日志表';
-- +goose StatementEnd

-- 7. 代取订单表
-- +goose StatementBegin
CREATE TABLE `proxy_orders` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '订单ID',
  `station_id` bigint NOT NULL COMMENT '驿站ID',
  `parcel_id` bigint NOT NULL COMMENT '关联包裹ID',
  `publisher_id` bigint NOT NULL COMMENT '发布者ID',
  `taker_id` bigint COMMENT '跑腿员ID',
  `reward_amount` decimal(10,2) NOT NULL COMMENT '悬赏金额(元)',
  `temp_pickup_code` varchar(10) COMMENT '临时取件码',
  `deadline` datetime NOT NULL COMMENT '取件截止时间',
  `status` tinyint DEFAULT 1 COMMENT '1-待接单, 2-配送中, 3-已完成, 4-已取消, 5-超时未接, 6-取件失败',
  `cancel_reason` varchar(255) COMMENT '取消原因',
  `delivery_time` datetime COMMENT '送达时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_parcel (`parcel_id`),
  INDEX idx_publisher (`publisher_id`),
  INDEX idx_taker (`taker_id`),
  INDEX idx_status (`status`),
  INDEX idx_station (`station_id`),
  INDEX idx_created_at (`created_at`),
  INDEX idx_station_status (`station_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='代取订单表';
-- +goose StatementEnd

-- 8. 货架布局表
-- +goose StatementBegin
CREATE TABLE `shelf_layout` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '货架ID',
  `station_id` bigint NOT NULL COMMENT '驿站ID',
  `shelf_code` varchar(20) NOT NULL COMMENT '货架编号',
  `row_num` int NOT NULL COMMENT '排数',
  `col_num` int NOT NULL COMMENT '列数',
  `current_capacity` int DEFAULT 0 COMMENT '当前占用',
  `max_capacity` int NOT NULL COMMENT '最大容量',
  `version` int DEFAULT 0 COMMENT '乐观锁版本号',
  `remark` varchar(255) COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_station_shelf (`station_id`, `shelf_code`),
  INDEX idx_station (`station_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='货架布局表';
-- +goose StatementEnd

-- 9. 通知记录表
-- +goose StatementBegin
CREATE TABLE `notifications` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '通知ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `parcel_id` bigint COMMENT '关联包裹ID',
  `title` varchar(100) NOT NULL COMMENT '通知标题',
  `content` text NOT NULL COMMENT '通知内容',
  `type` tinyint NOT NULL COMMENT '1-入库,2-催取,3-代取状态,4-系统',
  `is_read` tinyint DEFAULT 0 COMMENT '0-未读, 1-已读',
  `send_status` tinyint DEFAULT 0 COMMENT '0-待发送, 1-已发送, 2-发送失败',
  `channel` tinyint DEFAULT 1 COMMENT '1-微信, 2-短信',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user (`user_id`),
  INDEX idx_parcel (`parcel_id`),
  INDEX idx_send_status (`send_status`),
  INDEX idx_user_read (`user_id`, `is_read`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知记录表';
-- +goose StatementEnd

-- 10. 快递公司字典表
-- +goose StatementBegin
CREATE TABLE `courier_companies` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '快递公司ID',
  `company_name` varchar(50) NOT NULL COMMENT '公司名称',
  `company_code` varchar(20) NOT NULL COMMENT '快递100编码',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_company_name (`company_name`),
  UNIQUE KEY uk_company_code (`company_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='快递公司字典表';
-- +goose StatementEnd

-- 11. 操作日志表
-- +goose StatementBegin
CREATE TABLE `operation_logs` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint COMMENT '操作人ID',
  `user_type` tinyint COMMENT '1-管理员, 2-用户, 3-系统',
  `module` varchar(50) COMMENT '模块',
  `action` varchar(50) COMMENT '操作',
  `target_id` bigint COMMENT '操作对象ID',
  `request_params` text COMMENT '请求参数(JSON)',
  `ip` varchar(45) COMMENT 'IP地址',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user (`user_id`),
  INDEX idx_module_action (`module`, `action`),
  INDEX idx_created (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志表';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `operation_logs`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `courier_companies`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `notifications`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `shelf_layout`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `proxy_orders`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `pickup_logs`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `parcels`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `runner_applications`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `admins`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `users`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `stations`;
-- +goose StatementEnd
