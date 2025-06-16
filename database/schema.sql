-- --------------------------------------------------------
-- Hôte:                         127.0.0.1
-- Version du serveur:           9.1.0 - MySQL Community Server - GPL
-- SE du serveur:                Win64
-- HeidiSQL Version:             12.5.0.6677
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


-- Listage de la structure de la base pour forum
CREATE DATABASE IF NOT EXISTS `forum` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `forum`;

-- Listage de la structure de la table forum. categories
CREATE TABLE IF NOT EXISTS `categories` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL,
  `description` text COLLATE utf8mb4_general_ci,
  `color` varchar(7) COLLATE utf8mb4_general_ci DEFAULT '#007bff',
  `icon` varchar(50) COLLATE utf8mb4_general_ci DEFAULT 'book',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=MyISAM AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Listage des données de la table forum.categories : 7 rows
DELETE FROM `categories`;
/*!40000 ALTER TABLE `categories` DISABLE KEYS */;
INSERT INTO `categories` (`id`, `name`, `description`, `color`, `icon`, `created_at`) VALUES
	(1, 'Mathématiques', 'Aide en mathématiques, algèbre, géométrie', '#28a745', 'calculator', '2025-06-14 10:09:43'),
	(2, 'Français', 'Grammaire, littérature, rédaction', '#dc3545', 'feather', '2025-06-14 10:09:43'),
	(3, 'Sciences', 'Physique, chimie, biologie', '#17a2b8', 'atom', '2025-06-14 10:09:43'),
	(4, 'Histoire-Géographie', 'Histoire, géographie, éducation civique', '#ffc107', 'globe', '2025-06-14 10:09:43'),
	(5, 'Langues', 'Anglais, espagnol, allemand, etc.', '#6f42c1', 'message-circle', '2025-06-14 10:09:43'),
	(6, 'Informatique', 'Programmation, algorithmique', '#fd7e14', 'code', '2025-06-14 10:09:43'),
	(7, 'Philosophie', 'Réflexions philosophiques', '#6c757d', 'lightbulb', '2025-06-14 10:09:43');
/*!40000 ALTER TABLE `categories` ENABLE KEYS */;

-- Listage de la structure de la table forum. comments
CREATE TABLE IF NOT EXISTS `comments` (
  `id` int NOT NULL AUTO_INCREMENT,
  `post_id` int NOT NULL,
  `content` text COLLATE utf8mb4_general_ci NOT NULL,
  `user_id` int NOT NULL,
  `parent_id` int DEFAULT NULL,
  `is_solution` tinyint(1) DEFAULT '0',
  `likes_count` int DEFAULT '0',
  `dislikes_count` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `parent_id` (`parent_id`),
  KEY `idx_comments_post_id` (`post_id`),
  KEY `idx_comments_user_id` (`user_id`)
) ENGINE=MyISAM AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. comment_votes
CREATE TABLE IF NOT EXISTS `comment_votes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `comment_id` int NOT NULL,
  `user_id` int NOT NULL,
  `vote_type` enum('like','dislike') COLLATE utf8mb4_general_ci NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_comment_vote` (`comment_id`,`user_id`),
  KEY `user_id` (`user_id`),
  KEY `idx_votes_comment` (`comment_id`)
) ENGINE=MyISAM AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. images
CREATE TABLE IF NOT EXISTS `images` (
  `id` int NOT NULL AUTO_INCREMENT,
  `filename` varchar(100) COLLATE utf8mb4_general_ci NOT NULL,
  `original_name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL,
  `content_type` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `size_bytes` int NOT NULL,
  `width` int DEFAULT '0',
  `height` int DEFAULT '0',
  `post_id` int DEFAULT NULL,
  `comment_id` int DEFAULT NULL,
  `user_id` int NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `filename` (`filename`),
  KEY `idx_post_images` (`post_id`),
  KEY `idx_comment_images` (`comment_id`),
  KEY `idx_user_images` (`user_id`),
  KEY `idx_filename` (`filename`)
) ENGINE=MyISAM AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. moderation_logs
CREATE TABLE IF NOT EXISTS `moderation_logs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `moderator_id` int NOT NULL,
  `action_type` enum('ban_user','unban_user','delete_post','delete_comment','lock_post','unlock_post','pin_post','unpin_post','close_post','reopen_post','archive_post','unarchive_post') COLLATE utf8mb4_general_ci NOT NULL,
  `target_type` enum('user','post','comment') COLLATE utf8mb4_general_ci NOT NULL,
  `target_id` int NOT NULL,
  `reason` text COLLATE utf8mb4_general_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `moderator_id` (`moderator_id`)
) ENGINE=MyISAM AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. posts
CREATE TABLE IF NOT EXISTS `posts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(255) COLLATE utf8mb4_general_ci NOT NULL,
  `content` text COLLATE utf8mb4_general_ci NOT NULL,
  `user_id` int NOT NULL,
  `category_id` int NOT NULL,
  `status` enum('open','closed','archived') COLLATE utf8mb4_general_ci DEFAULT 'open',
  `is_solved` tinyint(1) DEFAULT '0',
  `is_pinned` tinyint(1) DEFAULT '0',
  `is_locked` tinyint(1) DEFAULT '0',
  `views_count` int DEFAULT '0',
  `likes_count` int DEFAULT '0',
  `dislikes_count` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_posts_category` (`category_id`),
  KEY `idx_posts_user_id` (`user_id`),
  KEY `idx_posts_created_at` (`created_at`),
  KEY `idx_posts_likes` (`likes_count`),
  KEY `idx_posts_status` (`status`)
) ENGINE=MyISAM AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. post_tags
CREATE TABLE IF NOT EXISTS `post_tags` (
  `post_id` int NOT NULL,
  `tag_id` int NOT NULL,
  PRIMARY KEY (`post_id`,`tag_id`),
  KEY `tag_id` (`tag_id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. post_votes
CREATE TABLE IF NOT EXISTS `post_votes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `post_id` int NOT NULL,
  `user_id` int NOT NULL,
  `vote_type` enum('like','dislike') COLLATE utf8mb4_general_ci NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_post_vote` (`post_id`,`user_id`),
  KEY `user_id` (`user_id`),
  KEY `idx_votes_post` (`post_id`)
) ENGINE=MyISAM AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. reports
CREATE TABLE IF NOT EXISTS `reports` (
  `id` int NOT NULL AUTO_INCREMENT,
  `reporter_id` int NOT NULL,
  `reported_type` enum('post','comment','user') COLLATE utf8mb4_general_ci NOT NULL,
  `reported_id` int NOT NULL,
  `reason` varchar(255) COLLATE utf8mb4_general_ci NOT NULL,
  `description` text COLLATE utf8mb4_general_ci,
  `status` enum('pending','reviewed','resolved','dismissed') COLLATE utf8mb4_general_ci DEFAULT 'pending',
  `moderator_id` int DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `resolved_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `reporter_id` (`reporter_id`),
  KEY `moderator_id` (`moderator_id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. roles
CREATE TABLE IF NOT EXISTS `roles` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `description` text COLLATE utf8mb4_general_ci,
  `permissions` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=MyISAM AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Listage des données de la table forum.roles : 4 rows
DELETE FROM `roles`;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;
INSERT INTO `roles` (`id`, `name`, `description`, `permissions`) VALUES
	(1, 'utilisateur', 'Utilisateur standard', '["read", "comment", "post"]'),
	(2, 'prof', 'Professeur', '["read", "comment", "post", "help_students", "verify_answers"]'),
	(3, 'moderateur', 'Modérateur', '["read", "comment", "post", "moderate", "delete_posts", "ban_users"]'),
	(4, 'administrateur', 'Administrateur', '["all"]');
/*!40000 ALTER TABLE `roles` ENABLE KEYS */;

-- Listage de la structure de la table forum. tags
CREATE TABLE IF NOT EXISTS `tags` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `color` varchar(7) COLLATE utf8mb4_general_ci DEFAULT '#6c757d',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=MyISAM AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. users
CREATE TABLE IF NOT EXISTS `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `email` varchar(100) COLLATE utf8mb4_general_ci NOT NULL,
  `password` varchar(255) COLLATE utf8mb4_general_ci NOT NULL,
  `role_id` int DEFAULT '1',
  `is_banned` tinyint(1) DEFAULT '0',
  `ban_reason` text COLLATE utf8mb4_general_ci,
  `banned_until` timestamp NULL DEFAULT NULL,
  `avatar` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `bio` text COLLATE utf8mb4_general_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `avatar_filename` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `last_login` timestamp NULL DEFAULT NULL,
  `profile_visibility` enum('public','private') COLLATE utf8mb4_general_ci DEFAULT 'public',
  `date_inscription` date DEFAULT NULL,
  `location` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_users_role` (`role_id`),
  KEY `idx_users_avatar` (`avatar_filename`),
  KEY `idx_users_visibility` (`profile_visibility`)
) ENGINE=MyISAM AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

-- Listage de la structure de la table forum. user_stats
CREATE TABLE IF NOT EXISTS `user_stats` (
  `user_id` int NOT NULL,
  `posts_count` int DEFAULT '0',
  `comments_count` int DEFAULT '0',
  `solutions_given` int DEFAULT '0',
  `solutions_received` int DEFAULT '0',
  `likes_received_posts` int DEFAULT '0',
  `likes_received_comments` int DEFAULT '0',
  `total_views_posts` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`),
  KEY `idx_user_stats_posts` (`posts_count`),
  KEY `idx_user_stats_solutions` (`solutions_given`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Les données exportées n'étaient pas sélectionnées.

/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
