
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for gb_configs
-- ----------------------------
DROP TABLE IF EXISTS `gb_configs`;
CREATE TABLE `gb_configs`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `namespace` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `field` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `type` enum('string','int','float','json') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `value` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `name`(`namespace`, `field`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic STORAGE MEMORY;

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
  `captcha_is_open` enum('Y','N') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'Y' COMMENT '是否开启验证码',
  `google_auth_code` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '谷歌验证码',
  `login_ip` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '最后一次登录IP',
  `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '注册时间',
  `updated_at` timestamp(0) DEFAULT NULL COMMENT '最后一次登录时间',
  `deleted_at` timestamp(0) DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username`) USING BTREE,
  UNIQUE INDEX `nickname`(`nickname`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of gb_admins
-- ----------------------------
INSERT INTO `gb_admins` VALUES (1, 'admin', 'admin', '123123', 'normal', 'Y', '', '127.0.0.1', '2019-11-16 16:31:49', '2019-11-16 16:31:47', NULL);

SET FOREIGN_KEY_CHECKS = 1;
