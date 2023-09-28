-- Add new schema named "example"
CREATE DATABASE `example` CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "authors" table
CREATE TABLE `example`.`authors` (`id` bigint NOT NULL AUTO_INCREMENT, `name` varchar(499) NOT NULL COMMENT "Full name of the author", `bio` text NULL, `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`id`), INDEX `authors_created_at_k` (`created_at`)) CHARSET utf8mb4 COLLATE utf8mb4_general_ci COMMENT "Author Table";
