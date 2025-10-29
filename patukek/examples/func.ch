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