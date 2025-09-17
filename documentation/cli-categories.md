# CLI Category Commands

## Listing

```bash
$ ./bctbackend [--no-color] category list
ID | Name
1  | Clothing 0-3 mos (50-56)
2  | Clothing 3-6 mos (56-62)
3  | Clothing 6-12 mos (68-80)
4  | Clothing 12-24 mos (86-92)
5  | Clothing 2-3 yrs (92-98)
6  | Clothing 4-6 yrs (104-116)
7  | Clothing 7-8 yrs (122-128)
8  | Clothing 9-10 yrs (128-140)
9  | Clothing 11-12 yrs (140-152)
10 | Books
11 | Toys
12 | Baby/Child Equipment
13 | Maternity
```

Lists categories.

## Item Counts by Category

```bash
$ ./bctbackend category count
ID | Name                         | Count
1  | Clothing 0-3 mos (50-56)     | 60
2  | Clothing 3-6 mos (56-62)     | 64
3  | Clothing 6-12 mos (68-80)    | 79
4  | Clothing 12-24 mos (86-92)   | 77
5  | Clothing 2-3 yrs (92-98)     | 67
6  | Clothing 4-6 yrs (104-116)   | 69
7  | Clothing 7-8 yrs (122-128)   | 59
8  | Clothing 9-10 yrs (128-140)  | 70
9  | Clothing 11-12 yrs (140-152) | 77
10 | Books                        | 589
11 | Toys                         | 568
12 | Baby/Child Equipment         | 0
13 | Maternity                    | 0
```

## Adding New Category

```bash
$ ./bctbackend --id <id> --name <name>

# Example
$ ./bctbackend --id 14 --name "Large Items"
```

No other category with same ID or name must exist, otherwise error.
