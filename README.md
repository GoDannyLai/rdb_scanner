# 简介
	扫描redis RDB文件， 找出大key. 注意不支持版本大于或者等于4.0.0的redis， 此时可使用https://github.com/GoDannyLai/redis_scanner
    线上遇到redis CPU高与网卡带宽跑满的情况， 很明显的bigkey问题，使用类似的工具来分析， 150MB的RDB需要1个小时才出结果。
	而使用rdb_scanner只需要63秒

    生成的bigkey报告为CSV格式：
![csv](https://github.com/GoDannyLai/rdb_scanner/raw/master/misc/img/bigkeys_csv.png)

    使用很简单，全部就6个参数：
        ./rdb_scanner.exe -b 1024 -o tmp/bigkeys_6380.csv -S -l 10 -t 3 tmp/dump6378.rdb
		上述命令以3个线程分析dump6378.rdb文件中大于1024bytes的KEY， 以CSV格式把最大的10个key输出到bigkeys_6380.csv的文件中
    