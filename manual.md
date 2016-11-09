# dbscale-sandbox manual

## Directory Structure

```
└── sandboxes
    ├── dbscale                             //dbscale path be installed
    ├── auth_mysql_sandbox                  //auth_MySQL is used as the login authentication
    │   ├── master
    │   │   ├── data                        //datadir
    │   │   ├── tmp                         //tmpdir
    │   │   ├── my.sandbox.cnf              //MySQL configure
    │   │   ├── mysql_sandbox3210.sock      //socket
    │   │   ├── send_kill                   //kill mysqld
    │   │   ├── start                       //start MySQL
    │   │   ├── stop                        //stop MySQL
    │   │   └── use                         //login MySQL
    │   └── slave
    │       ├── data
    │       ├── tmp
    │       └── ...
    ├── rep_mysql_sandbox1                  //MySQL partition 1
    │   ├── master
    │   │   ├── data
    │   │   ├── tmp
    │   │   └── ...
    │   └── slave
    │       ├── data
    │       ├── tmp
    │       └── ...
    ├── rep_mysql_sandbox2                  //MySQL partition 2
    │   ├── master
    │   │   ├── data
    │   │   ├── tmp
    │   │   └── ...
    │   └── slave
    │       ├── data
    │       ├── tmp
    │       └── ...
    ├── mysql.grants                        //grants sql
    ├── loginDBScale                        //login dbscale
    ├── dbscale-start.sh                    //start dbscale
    ├── dbscale-stop.sh                     //stop dbscale
    ├── startallmysql                       //start MySQL
    ├── stopallmysql                        //stop MySQL
    ├── startall                            //start dbscale-sandbox
    └── stopall                             //stop dbscale-sandbox
```


## start & stop

### dbscale-sandbox
```
./startall      //start dbscale-sandbox
./stopall       //stop dbscale-sandbox
```

### dbscale

```
./dbscale-start.sh  //start dbscale
./dbscale-stop.sh   //stop dbscale
```

### MySQL

```
./startallmysql     //start MySQL
./stopallmysql      //stop MySQL
```
## login dbscale

```
./loginDBScale
```

## test dbscale

```
$ ./loginDBScale test
mysql> show tables;
+----------------+
| Tables_in_test |
+----------------+
| part_tb01      |
| part_tb02      |
+----------------+
2 rows in set (0.01 sec)

mysql> select * from part_tb01;
+-----+------+---------------------------------+
| id  | c1   | c2                              |
+-----+------+---------------------------------+
|   1 |    1 | hello world.                    |
|   2 |    2 | welecome to dbscale.            |
|   3 |    3 | this is a demo partition table. |
|   8 |    8 | this is test text8              |
|   9 |    9 | this is test text9              |
|  10 |   10 | this is test text10             |
|  11 |   11 | this is test text11             |
|  12 |   12 | this is test text12             |
|  13 |   13 | this is test text13             |
|  18 |   18 | this is test text18             |
|  19 |   19 | this is test text19             |
|  20 |   20 | this is test text20             |
|  21 |   21 | this is test text21             |
|  22 |   22 | this is test text22             |
|  23 |   23 | this is test text23             |
|  28 |   28 | this is test text28             |
|  29 |   29 | this is test text29             |
|  30 |   30 | this is test text30             |
|  31 |   31 | this is test text31             |
|  32 |   32 | this is test text32             |
|  33 |   33 | this is test text33             |
|  38 |   38 | this is test text38             |
|  39 |   39 | this is test text39             |
|  40 |   40 | this is test text40             |
|  41 |   41 | this is test text41             |
|  42 |   42 | this is test text42             |
|  43 |   43 | this is test text43             |
|  48 |   48 | this is test text48             |
|  49 |   49 | this is test text49             |
|  50 |   50 | this is test text50             |
|  51 |   51 | this is test text51             |
|  52 |   52 | this is test text52             |
|  53 |   53 | this is test text53             |
|  58 |   58 | this is test text58             |
|  59 |   59 | this is test text59             |
|  60 |   60 | this is test text60             |
|  61 |   61 | this is test text61             |
|  62 |   62 | this is test text62             |
|  63 |   63 | this is test text63             |
|  68 |   68 | this is test text68             |
|  69 |   69 | this is test text69             |
|  70 |   70 | this is test text70             |
|  71 |   71 | this is test text71             |
|  72 |   72 | this is test text72             |
|  73 |   73 | this is test text73             |
|   4 |    4 | plz try and have fun.           |
|   5 |    5 | this is test text5              |
|   6 |    6 | this is test text6              |
|   7 |    7 | this is test text7              |
|  14 |   14 | this is test text14             |
|  15 |   15 | this is test text15             |
|  16 |   16 | this is test text16             |
|  17 |   17 | this is test text17             |
|  24 |   24 | this is test text24             |
|  25 |   25 | this is test text25             |
|  26 |   26 | this is test text26             |
|  27 |   27 | this is test text27             |
|  34 |   34 | this is test text34             |
|  35 |   35 | this is test text35             |
|  36 |   36 | this is test text36             |
|  37 |   37 | this is test text37             |
|  44 |   44 | this is test text44             |
|  45 |   45 | this is test text45             |
|  46 |   46 | this is test text46             |
|  47 |   47 | this is test text47             |
|  54 |   54 | this is test text54             |
|  55 |   55 | this is test text55             |
|  56 |   56 | this is test text56             |
|  57 |   57 | this is test text57             |
|  64 |   64 | this is test text64             |
|  65 |   65 | this is test text65             |
|  66 |   66 | this is test text66             |
|  67 |   67 | this is test text67             |
|  74 |   74 | this is test text74             |
|  75 |   75 | this is test text75             |
|  76 |   76 | this is test text76             |
|  77 |   77 | this is test text77             |
|  84 |   84 | this is test text84             |
|  85 |   85 | this is test text85             |
|  86 |   86 | this is test text86             |
|  87 |   87 | this is test text87             |
|  94 |   94 | this is test text94             |
|  95 |   95 | this is test text95             |
|  96 |   96 | this is test text96             |
|  97 |   97 | this is test text97             |
|  78 |   78 | this is test text78             |
|  79 |   79 | this is test text79             |
|  80 |   80 | this is test text80             |
|  81 |   81 | this is test text81             |
|  82 |   82 | this is test text82             |
|  83 |   83 | this is test text83             |
|  88 |   88 | this is test text88             |
|  89 |   89 | this is test text89             |
|  90 |   90 | this is test text90             |
|  91 |   91 | this is test text91             |
|  92 |   92 | this is test text92             |
|  93 |   93 | this is test text93             |
|  98 |   98 | this is test text98             |
|  99 |   99 | this is test text99             |
| 100 |  100 | this is test text100            |
+-----+------+---------------------------------+
100 rows in set (0.01 sec)
```
