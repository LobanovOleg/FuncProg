# FuncProg

# Отчет по курсу "Функциональное программирование"

### Студенты: 

| ФИО           | Группа      | Роль в проекте                     | Оценка       |
|---------------|-------------|------------------------------------|--------------|
| Лобанов Олег  | М8О-412Б-22 | Сделал курсовой проект             |              |


> *Комментарии проверяющих (обратите внимание, что более подробные комментарии возможны непосредственно в репозитории по тексту программы)*

## Пример программ на языке программирования "patukek"

### hello.ptk

Для запуска:

```
$ go run main.go examples/hello.ptk
```

Программа:

```
println("Hello World")

```

Результат:

```
Hello World
```

### calc.ptk

Для запуска:

```
$ go run main.go examples/calc.ptk
```

Программа:

```
a = 5
b = 6

c = a + b
println(c)

d = a - b
println(d)

e = a * b
println(e)

f = e / b
println(f)

g = e % c
println(g)

h = (a + b) * c - d % g
println(h)

```

Результат:

```
11
-1
30
5
8
122
```

### cond.ptk

Для запуска:

```
$ go run main.go examples/cond.ptk
```

Программа:

```
year = 2023
if year == 2023 {
	println("2023")
}

year = 2020
if year >= 2023 {
    println("more or equal")
} else {
    println("less")
}

year = 2025
if year <= 2023 {
    println("less of equal")
} else {
    println("more")
}

year = 2020
if year > 2023 {
    println("more")
} else {
    println("less or equal")
}

year = 2025
if year < 2023 {
    println("less")
} else {
    println("more or equal")
}

year = 2020
if year > 2000 && year < 2100 {
    println("21st century")
}

year = 2020
if (year == 2016 || year == 2020 || year == 2024) && (year != 2100) {
    println("leap year")
}

```

Результат:

```
2023
less
more
less or equal
more or equal
21st century
leap year
```

### list.ptk

Для запуска:

```
$ go run main.go examples/list.ptk
```

Программа:

```
lst = [0, 1, 2, 3, 4]

println(lst)
println(len(lst))

```

Результат:

```
[0, 1, 2, 3, 4]
5
```

### func.ptk

Для запуска:

```
$ go run main.go examples/func.ptk
```

Программа:

```
min = patukek(a, b) {
    if a < b {
        return a
    }
    b
}

max = patukek(a, b) {
    if a > b {
        return a
    }
    b
}

comp = patukek(a, b, f) {
    f(a, b)
}

println(comp(1, 2, min))
println(comp(1, 2, max))

```

Результат:

```
1
2
```

### loop.ptk

Для запуска:

```
$ go run main.go examples/loop.ptk
```

Программа:

```
loop = patukek(times, function) {
	if times > 0 {
		function()
		loop(times-1, function)
	}
}

loop(5, patukek() { println("Hello World") })

```

Результат:

```
Hello World
Hello World
Hello World
Hello World
Hello World
```

### rec.ptk

Для запуска:

```
$ go run main.go examples/rec.ptk
```

Программа:

```
fib = patukek(n) {
	if n < 2 {
		return n
	}
	fib(n-1) + fib(n-2)
}

println(fib(4))
println(fib(5))
println(fib(6))

_tail_fib = patukek(n, acc1, acc2) {
    if n < 2 {
        return acc1
    }
    _tail_fib(n-1, acc1 + acc2, acc1)
}

tail_fib = patukek(n) {
    _tail_fib(n, 1, 0)
}

println(tail_fib(4))
println(tail_fib(5))
println(tail_fib(6))

```

Результат:

```
3
5
8
3
5
8
```

### error.ptk

Для запуска:

```
$ go run main.go examples/error.ptk
```

Программа:

```
a = 5
b = 6

println(a + b + c)

```

Результат:

```
PATUKEK! error in file examples/error.ptk at line 4:
    println(a + b + c)
                    ^
undefined variable c
```

## Синтаксическое дерево

AST представляет собой структуру данных, которая отражает синтаксическую 
структуру программы, используя функции и их аргументы в качестве основных 
элементов. В языке программирования "patukek" синтаксическое дерево состоит из 
узлов, представляющих выражения или функции, и их аргументов. AST позволяет 
программе анализировать и работать с синтаксической структурой программы на 
более абстрактном уровне. Он может использоваться для проверки типов, 
оптимизации кода, генерации промежуточного представления и других операций, 
связанных с анализом программы.

## Интерпретатор

У меня получился транслируемый язык программирования. При подаче в него файла с 
программой на языке "patukek" программа сначала транслируется в байт-код, затем 
этот байт-код выполняется виртуальной машиной языка "patukek". Всё это происходит в 
рамках единственного выполнения программы, то есть в плане выполнения программ 
язык программирования "patukek" похож на интерпретируемые языки программирования, 
хотя в теории возможно сохранение файлов с байт-кодом и последующее их 
выполнение без транслирования.

## Какие фишки вы реализовали

* [ ] Именованные переменные
* [x] Рекурсия
* [ ] Ленивые вычисления
* [x] Функции
* [ ] Замыкания

## Getting Started with "patukek"

В языке программирования "patukek" реализованы два типа данных: целое число и 
строка. 

Над целыми числами возможны все основные операции: 

* сложение `+`
* вычитание `-`
* умножение `*`
* целочисленное деление `/`
* взятие остатка `%`

Синтаксис инициализации переменной и присваивания ей значения соответствует 
такому в языках Python и Ruby, блоки кода задаются фигурными скобками, как во 
всех си-подобных языках. В языке "patukek" также есть условные операторы.

Основной элемент любого функционального языка программирования — это функция. 
Функции в моем языке создаются с помощью ключевого слова `patukek`, могут присваиваться как переменные и 
передаваться в качестве аргументов другим функциям.