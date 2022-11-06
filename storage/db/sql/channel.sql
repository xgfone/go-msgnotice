CREATE TABLE `channel` (
    -- Common Information
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'the primary key',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'the created time',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'the last updated time',
    `deleted_at` DATETIME NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'the deleted time; if zero, not deleted',

    -- Channel Information
    `channel_name` VARCHAR(64) NOT NULL COMMENT 'the channel name',
    `driver_name` VARCHAR(64) NOT NULL COMMENT 'the driver name',
    `driver_type` VARCHAR(16) NOT NULL DEFAULT '' COMMENT 'the driver type',
    `driver_conf` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'the driver config',

    PRIMARY KEY `pk_channel`(`id`),
    INDEX `idx_channel_cname` (`deleted_at`, `channel_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
