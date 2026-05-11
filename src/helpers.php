<?php

if (!function_exists('slugify')) {
    function slugify(string $str): string
    {
        $str = mb_strtolower(trim($str));
        $str = preg_replace('/[^a-z0-9]+/', '-', $str);
        return trim($str, '-') ?: 'image';
    }
}
