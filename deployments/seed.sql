INSERT INTO users (
    id,
    username,
    display_name,
    bio,
    timezone,
    role,
    status,
    created_at,
    updated_at
) VALUES (
    '11111111-1111-4111-8111-111111111111',
    'hachi',
    'Первый пользователь',
    'Тестовый профиль локальной среды HachiSocial.',
    'Europe/Moscow',
    'user',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;
