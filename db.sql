CREATE TABLE `quotes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `contract_id` int(11) NOT NULL,
  `session_volume` int(11) NOT NULL,
  `open_interest` int(11) NOT NULL,
  `opening_price` float DEFAULT NULL,
  `high_price` float DEFAULT NULL,
  `settlement_price` float DEFAULT NULL,
  `low_price` float DEFAULT NULL,
  `bid_price` float DEFAULT NULL,
  `bid_size` int(11) DEFAULT NULL,
  `ask_price` float DEFAULT NULL,
  `ask_size` int(11) DEFAULT NULL,
  `trade_price` float DEFAULT NULL,
  `trade_size` int(11) DEFAULT NULL,
  `timestamp` bigint(20) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`)
)