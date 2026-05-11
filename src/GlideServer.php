<?php

namespace App;

use League\Glide\ServerFactory;
use League\Glide\Server;

class GlideServer
{
    public static function create(): Server
    {
        return ServerFactory::create([
            'source'         => dirname(__DIR__) . '/storage/source',
            'cache'          => dirname(__DIR__) . '/storage/cache',
            'max_image_size' => 4000 * 4000,
        ]);
    }
}
