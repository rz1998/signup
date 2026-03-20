#!/usr/bin/env python3
"""
报名系统 Logo & 图标生成工具
"""
import os
from PIL import Image, ImageDraw, ImageFont

# 确保目录存在
assets_dir = '/home/rz1998/workspace/signup/miniprogram/assets'
os.makedirs(assets_dir, exist_ok=True)
os.makedirs(f'{assets_dir}/logos', exist_ok=True)
os.makedirs(f'{assets_dir}/icons', exist_ok=True)

# 颜色定义
PRIMARY_BLUE = '#1890FF'
DARK_BLUE = '#0050B3'
SUCCESS_GREEN = '#52C41A'
WARNING_ORANGE = '#FA8C16'
ERROR_RED = '#FF4D4F'
PURPLE = '#722ED1'
CYAN = '#13C2C2'

def create_rounded_rectangle(draw, size, radius, fill_color):
    """创建圆角矩形"""
    width, height = size
    # 画圆角矩形
    draw.rounded_rectangle([(0, 0), (width-1, height-1)], radius=radius, fill=fill_color)

def hex_to_rgb(hex_color):
    """HEX转RGB"""
    hex_color = hex_color.lstrip('#')
    return tuple(int(hex_color[i:i+2], 16) for i in (0, 2, 4))

def create_icon(name, size, icon_svg_path, bg_color):
    """创建图标"""
    img = Image.new('RGBA', (size, size), (255, 255, 255, 0))
    draw = ImageDraw.Draw(img)
    
    # 背景
    create_rounded_rectangle(draw, (size, size), size//5, bg_color)
    
    # 简单的图形化图标
    center = size // 2
    icon_size = size // 3
    left = center - icon_size // 2
    right = center + icon_size // 2
    top = center - icon_size // 2
    bottom = center + icon_size // 2
    
    # 图标颜色
    rgb = hex_to_rgb(bg_color)
    lighter = tuple(min(255, c + 40) for c in rgb)
    
    # 画不同形状的图标
    if 'activity' in name or 'form' in name:
        # 表格图标
        draw.rectangle([left, top, right, bottom], fill=(255,255,255,200))
        line_y = center
        draw.line([(left, line_y), (right, line_y)], fill=bg_color, width=max(1, size//20))
        line_x = center
        draw.line([(line_x, top), (line_x, bottom)], fill=bg_color, width=max(1, size//20))
    elif 'user' in name or 'users' in name:
        # 用户图标
        head_size = icon_size // 3
        draw.ellipse([center-head_size//2, top+icon_size//4, center+head_size//2, top+icon_size//4+head_size], fill=(255,255,255,200))
        body_top = top + icon_size // 4 + head_size
        draw.ellipse([left, body_top, right, bottom], fill=(255,255,255,200))
    elif 'share' in name:
        # 分享图标 - 三个圆点
        dot_size = icon_size // 5
        for dx, dy in [(-icon_size//4, 0), (0, 0), (icon_size//4, 0)]:
            draw.ellipse([center+dx-dot_size//2, center+dy-dot_size//2, center+dx+dot_size//2, center+dy+dot_size//2], fill=(255,255,255,200))
    elif 'data' in name or 'chart' in name:
        # 柱状图图标
        bar_width = icon_size // 5
        for i, h in enumerate([0.4, 0.7, 0.5, 0.9, 0.6]):
            x = left + i * (bar_width + 2)
            bar_h = int(icon_size * h)
            draw.rectangle([x, bottom - bar_h, x + bar_width, bottom], fill=(255,255,255,200))
    elif 'setting' in name or 'config' in name:
        # 齿轮图标
        draw.ellipse([left + icon_size//4, top + icon_size//4, right - icon_size//4, bottom - icon_size//4], fill=(255,255,255,200))
        draw.ellipse([center-2, center-2, center+2, center+2], fill=bg_color)
    elif 'company' in name or 'building' in name:
        # 建筑/公司图标
        draw.rectangle([left, center, right, bottom], fill=(255,255,255,200))
        # 屋顶
        draw.polygon([(left-2, center), (center, top), (right+2, center)], fill=(255,255,255,200))
    elif 'branch' in name or 'org' in name:
        # 组织架构图标
        draw.ellipse([center-3, top, center+3, top+6], fill=(255,255,255,200))
        draw.line([(center, top+6), (center, center)], fill=(255,255,255,200), width=2)
        for child_x in [left, center, right]:
            draw.ellipse([child_x-3, center, child_x+3, center+6], fill=(255,255,255,200))
            draw.line([(center, center+3), (child_x, center)], fill=(255,255,255,200), width=2)
    else:
        # 默认图标
        draw.ellipse([left, top, right, bottom], fill=(255,255,255,200))
    
    return img

def create_logo(size=(200, 60)):
    """创建Logo主图"""
    img = Image.new('RGBA', size, (255, 255, 255, 0))
    draw = ImageDraw.Draw(img)
    
    # 图标区域
    icon_size = 48
    icon_img = create_icon('activity', icon_size, None, PRIMARY_BLUE)
    img.paste(icon_img, (0, (size[1] - icon_size) // 2), icon_img)
    
    # 文字
    try:
        font = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf', 24)
        font_cn = ImageFont.truetype('/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc', 24)
    except:
        font = ImageFont.load_default()
        font_cn = font
    
    # 绘制文字
    text = "报名系统"
    draw.text((icon_size + 10, (size[1] - 24) // 2), text, font=font_cn, fill='#1a1a1a')
    
    return img

def create_logo_compact(size=(100, 40)):
    """创建紧凑版Logo"""
    img = Image.new('RGBA', size, (255, 255, 255, 0))
    draw = ImageDraw.Draw(img)
    
    icon_size = 32
    icon_img = create_icon('form', icon_size, None, PRIMARY_BLUE)
    img.paste(icon_img, (0, (size[1] - icon_size) // 2), icon_img)
    
    try:
        font = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf', 16)
    except:
        font = ImageFont.load_default()
    
    draw.text((icon_size + 6, (size[1] - 16) // 2), "SignUp", font=font, fill='#333333')
    
    return img

def create_placeholder_image(filename, size, text, bg_color='#e6e6e6'):
    """创建占位图"""
    img = Image.new('RGB', size, bg_color)
    draw = ImageDraw.Draw(img)
    
    try:
        font = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf', size[0] // 10)
    except:
        font = ImageFont.load_default()
    
    # 添加文字
    text_y = size[1] // 2 - 20
    draw.text((10, text_y), text, font=font, fill='#999999')
    
    img.save(filename)
    print(f"Created: {filename}")

def main():
    print("Generating logo and icons for 报名系统...")
    
    # 创建占位图
    create_placeholder_image(f'{assets_dir}/images/default-cover.png', (750, 400), 'Cover Image')
    
    # 创建TabBar图标 (48x48)
    tabbar_icons = [
        ('home.png', 'home'),
        ('home-active.png', 'home'),
        ('activity.png', 'activity'),
        ('activity-active.png', 'activity'),
        ('users.png', 'user'),
        ('users-active.png', 'user'),
        ('my.png', 'setting'),
        ('my-active.png', 'setting'),
    ]
    
    for filename, icon_type in tabbar_icons:
        size = 48
        is_active = 'active' in filename
        color = PRIMARY_BLUE if is_active else '#999999'
        img = create_icon(icon_type, size, None, color)
        img.save(f'{assets_dir}/icons/{filename}')
        print(f"Created: {filename}")
    
    # 创建功能图标 (64x64)
    feature_icons = [
        ('company.png', 'company', PRIMARY_BLUE),
        ('branch.png', 'branch', SUCCESS_GREEN),
        ('activity-icon.png', 'activity', PRIMARY_BLUE),
        ('user-icon.png', 'user', SUCCESS_GREEN),
        ('share-icon.png', 'share', WARNING_ORANGE),
        ('data-icon.png', 'data', PURPLE),
        ('setting-icon.png', 'setting', CYAN),
    ]
    
    for filename, icon_type, color in feature_icons:
        size = 64
        img = create_icon(icon_type, size, None, color)
        img.save(f'{assets_dir}/icons/{filename}')
        print(f"Created: {filename}")
    
    # 创建Logo
    logo = create_logo((240, 60))
    logo.save(f'{assets_dir}/logos/logo-main.png')
    print(f"Created: logos/logo-main.png")
    
    logo_dark = create_logo((240, 60))
    logo_dark.save(f'{assets_dir}/logos/logo-dark.png')
    print(f"Created: logos/logo-dark.png")
    
    logo_compact = create_logo_compact((120, 40))
    logo_compact.save(f'{assets_dir}/logos/logo-compact.png')
    print(f"Created: logos/logo-compact.png")
    
    # 创建头像占位图
    avatar = Image.new('RGB', (80, 80), '#e6e6e6')
    draw = ImageDraw.Draw(avatar)
    try:
        font = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf', 24)
    except:
        font = ImageFont.load_default()
    draw.text((20, 28), "User", font=font, fill='#999999')
    avatar.save(f'{assets_dir}/icons/default-avatar.png')
    print(f"Created: icons/default-avatar.png")
    
    print("\nAll assets generated successfully!")

if __name__ == '__main__':
    main()
