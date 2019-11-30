/*
 Navicat Premium Data Transfer

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 50725
 Source Host           : 127.0.0.3:3306
 Source Schema         : goblog

 Target Server Type    : MySQL
 Target Server Version : 50725
 File Encoding         : 65001

 Date: 21/11/2019 16:22:25
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for gb_admins
-- ----------------------------
DROP TABLE IF EXISTS `gb_admins`;
CREATE TABLE `gb_admins`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `nickname` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '昵称',
  `username` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '登录账号',
  `password` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '登录密码',
  `status` enum('locked','normal') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'normal' COMMENT '账号状态',
  `captcha_is_open` enum('Y','C') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'Y' COMMENT '是否开启验证码',
  `google_auth_code` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '谷歌验证码',
  `login_ip` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '最后一次登录IP',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '注册时间',
  `updated_at` timestamp(0) DEFAULT NULL COMMENT '最后一次登录时间',
  `deleted_at` timestamp(0) DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username`) USING BTREE,
  UNIQUE INDEX `nickname`(`nickname`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of gb_admins
-- ----------------------------
INSERT INTO `gb_admins` VALUES (1, '花开', 'admin', '$2a$10$X0VV5YWrmowEpiqnVxPk0e8VFQBmwWKrk.AIeFWUOgY8uPrS2iFcO', 'normal', 'Y', '', '127.0.0.1', '2019-11-21 09:13:41', '2019-11-16 16:31:47', NULL);

-- ----------------------------
-- Table structure for gb_admins_log
-- ----------------------------
DROP TABLE IF EXISTS `gb_admins_log`;
CREATE TABLE `gb_admins_log`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `login_uid` int(10) UNSIGNED NOT NULL,
  `login_username` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `visit_datetime` datetime(0) NOT NULL,
  `ip` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `action` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `msg` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `info` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `login_uid`(`login_uid`) USING BTREE,
  INDEX `login_username`(`login_username`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for gb_configs
-- ----------------------------
DROP TABLE IF EXISTS `gb_configs`;
CREATE TABLE `gb_configs`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `namespace` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `field` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `type` enum('string','int','float','json','time') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `value` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `name`(`namespace`, `field`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 7 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic STORAGE MEMORY;

-- ----------------------------
-- Records of gb_configs
-- ----------------------------
INSERT INTO `gb_configs` VALUES (1, 'admin', 'login_captcha', 'string', 'condition');
INSERT INTO `gb_configs` VALUES (2, 'admin', 'login_captcha_condition', 'json', '{\"pwd_errn\": 1, \"captcha_errn\": 1}');
INSERT INTO `gb_configs` VALUES (3, 'admin', 'login_captcha_config', 'json', ' {\r\n	\"Height\": 60,\r\n	\"Width\": 240,\r\n	\"Mode\": 1,\r\n	\"ComplexOfNoiseText\": 0,\r\n	\"ComplexOfNoiseDot\": 0,\r\n	\"IsUseSimpleFont\": true,\r\n	\"IsShowHollowLine\": false,\r\n	\"IsShowNoiseDot\": false,\r\n	\"IsShowNoiseText\": false,\r\n	\"IsShowSlimeLine\": false,\r\n	\"IsShowSineLine\": false,\r\n	\"CaptchaLen\": 6,\r\n	\"BgColor\": {\r\n		\"R\": 0,\r\n		\"G\": 0,\r\n		\"B\": 0,\r\n		\"A\": 1\r\n	}\r\n}');
INSERT INTO `gb_configs` VALUES (4, 'admin', 'login_malice_prevent', 'json', '{\"pwd_errn\": 3, \"lock_time\": 3600}');
INSERT INTO `gb_configs` VALUES (5, 'admin', 'login_counter_expire', 'time', '10800');
INSERT INTO `gb_configs` VALUES (6, 'admin', 'login_captcha_expire', 'time', '180');

SET FOREIGN_KEY_CHECKS = 1;
