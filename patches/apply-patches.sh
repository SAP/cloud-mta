for i in *.patch; do
    [ -f "$i" ] || break
    # the `git apply` must be executed in the root
    echo "applying $i"
    (cd .. && git apply patches/$i --unsafe-paths --verbose)
done