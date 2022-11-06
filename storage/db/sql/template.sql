CREATE TABLE `template` (
    -- Common Information
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'the primary key',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'the created time',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'the last updated time',
    `deleted_at` DATETIME NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'the deleted time; if zero, not deleted',

    -- Template Information
    `name` VARCHAR(64) NOT NULL COMMENT 'the template name',
    `tmpl` VARCHAR(512) NOT NULL COMMENT 'the template content',
    `args` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'the template arguments',

    PRIMARY KEY `pk_template`(`id`),
    INDEX `idx_template_name` (`deleted_at`, `name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
