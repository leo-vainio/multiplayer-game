import pygame
import os
pygame.font.init()
pygame.mixer.init()

WIDTH, HEIGHT = 1440, 900
WIN = pygame.display.set_mode((WIDTH, HEIGHT))
pygame.display.set_caption("Agar.io from Wish.com!")
pygame.display.set_icon(pygame.image.load(os.path.join('Assets', 'icon.png')))

WHITE = (255, 255, 255)
BLACK = (0,0,0)
YELLOW = (255, 255, 0)
RED = (255, 0, 0)

# food will be sent from server with their x and y -coordinate as the center of the circle, no radius specified
# maybe food will need to be sent with color since otherwise we would update it with random colors all the time
# maybe we can make some sort of optimization here.
# player will be sent with a bit more information, name, color, radius, food-size? (maybe same as radius tho)

def draw_window():
    WIN.fill(WHITE)

    pygame.draw.circle(WIN, (33, 44, 55), (400, 400), 50, 5)
    pygame.draw.circle(WIN, (33, 44, 55), (700, 500), 10, 0)

    pygame.draw.circle(WIN, (33, 66, 77), (1000, 400), 50, 0)
    pygame.draw.circle(WIN, (33, 44, 55), (1000, 400), 50, 5)

    pygame.draw.circle(WIN, (255, 44, 55), (1200, 500), 4, 0)

    pygame.display.update()

def read_players(amount):
    pass
    # (r, g, b) -> 3 bytes
    # center: (x, y) -> 4 bytes, 2 each (uint16)
    # radius -> 2 bytes
    # score -> 2 bytes
    # name -> string\n (varied) (gets displayed on the players ball(s))
  
def main():
    FPS = 60
    clock = pygame.time.Clock()
    running = True

    food_balls = []

    while running:
        clock.tick(FPS)

        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                running = False
                pygame.quit()
    
        draw_window()

if __name__ == "__main__":
    main()
