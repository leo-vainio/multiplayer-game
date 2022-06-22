import argparse
import os
import pygame
import socket
import struct
from random import randrange

pygame.init()
pygame.font.init()

parser = argparse.ArgumentParser(description='http server address')
parser.add_argument('--addr', default='127.0.0.1', help='Usage: --addr <ip>')
parser.add_argument('--port', type=int, default=8080, help='Usage: --port <port>')
args = parser.parse_args() 

s = socket.socket()
s.connect((args.addr, args.port))

WIDTH, HEIGHT = 1440, 900
WIN = pygame.display.set_mode((WIDTH, HEIGHT))
pygame.display.set_caption("Agar.IO from Wish.com!")
pygame.display.set_icon(pygame.image.load(os.path.join('Assets', 'icon.png')))

BACKGROUND = pygame.transform.scale(
    pygame.image.load(os.path.join('Assets', 'background.jpg')), (WIDTH, HEIGHT))

MENU_HEADER_FONT = pygame.font.SysFont('couriernew', 120)
MENU_NAME_FONT = pygame.font.SysFont('couriernew', 30)
MENU_HELP_FONT = pygame.font.SysFont('couriernew', 20)


WHITE = (255, 255, 255)
BLACK = (0, 0, 0)


# draw_menu draws all necessary components to the menu screen.
def draw_menu(name, color):
    WIN.blit(BACKGROUND, (0,0))
    pygame.draw.circle(WIN, color, (WIDTH/2, HEIGHT/2), 100, 0)

    header = MENU_HEADER_FONT.render("Agar.IO", 1, BLACK)
    WIN.blit(header, (WIDTH/2 - header.get_width()/2, HEIGHT/4 - header.get_height()/2))

    name_text = MENU_NAME_FONT.render(name, 1, BLACK)
    WIN.blit(name_text, (WIDTH/2 - name_text.get_width()/2, HEIGHT/2 - name_text.get_height()/2))

    text_field_width, text_field_height = 400, 40
    text_field = pygame.Rect(WIDTH/2 - text_field_width/2, HEIGHT - HEIGHT/4, text_field_width, text_field_height)
    pygame.draw.rect(WIN, BLACK, text_field, 2)

    text_field_header = MENU_NAME_FONT.render("Name:", 1, BLACK)
    WIN.blit(text_field_header, (WIDTH/2 - text_field_width/2 + 5, HEIGHT - HEIGHT/4 - text_field_header.get_height()))

    text_field_text = MENU_NAME_FONT.render(name, 1, BLACK)
    WIN.blit(text_field_text, (WIDTH/2 - text_field_width/2 + 5, HEIGHT - HEIGHT/4 + 5))

    help_text = MENU_HELP_FONT.render(
        "Type your username with keyboard, press <CTRL> to change color, press <ENTER> to play!", 1, BLACK)
    WIN.blit(help_text, (WIDTH/2 - help_text.get_width()/2, HEIGHT- help_text.get_height() - 10))

    pygame.display.update()


# random_color returns a random RGB color.
def random_color():
    return (randrange(255), randrange(255), randrange(255))


# handle_menu returns True if player information was successfully sent to server and False if player quit out of pygame.
def handle_menu():
    (R, G, B) = random_color()
    name = ""
    while True: 
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                pygame.quit()
                return False
            if event.type == pygame.KEYDOWN:
                if event.key == pygame.K_BACKSPACE:
                    name = name[:-1]
                if event.key == pygame.K_LCTRL or event.key == pygame.K_RCTRL:
                    (R, G, B) = random_color()
                if event.unicode.isalnum() and len(name) < 15:
                    name += event.unicode

        keys_pressed = pygame.key.get_pressed()
        if keys_pressed[pygame.K_RETURN] and name != "":
            s.send(R.to_bytes(1, 'little'))
            s.send(G.to_bytes(1, 'little'))
            s.send(B.to_bytes(1, 'little'))
            s.send((name + '\n').encode())
            return True

        draw_menu(name, (R, G, B))
        
# draw_player draws a single player onto the screen.
def draw_player(x, y, color, rad, name):
    pygame.draw.circle(WIN, color, (x, y), rad, 0)

    GAME_NAME_FONT = pygame.font.SysFont('couriernew', int(rad)//2)
    name_text = GAME_NAME_FONT.render(name[:-1], 1, BLACK)
    WIN.blit(name_text, (x - name_text.get_width()/2, y - name_text.get_height()/2))

def draw_food(x, y, color, rad):
    pygame.draw.circle(WIN, color, (x, y), rad, 0)

# read_and_draw_game recieves game data from server and draws the updated game onto the screen.
def recv_and_draw_game():
    WIN.blit(BACKGROUND, (0,0))

    status = int.from_bytes(s.recv(1), "little")
    num_food = int.from_bytes(s.recv(1), "little")
    for _ in range(num_food):
        x = int.from_bytes(s.recv(2), "little") # does not get read
        y = int.from_bytes(s.recv(2), "little")

        r = int.from_bytes(s.recv(1), "little")
        g = int.from_bytes(s.recv(1), "little")
        b = int.from_bytes(s.recv(1), "little")
        [rad] = struct.unpack('f', s.recv(4))

        draw_food(x, y, (r, g, b), rad)

    num_players = int.from_bytes(s.recv(1), "little")
    for _ in range(num_players):
        x = int.from_bytes(s.recv(2), "little")
        y = int.from_bytes(s.recv(2), "little")

        r = int.from_bytes(s.recv(1), "little")
        g = int.from_bytes(s.recv(1), "little")
        b = int.from_bytes(s.recv(1), "little")
    
        [rad] = struct.unpack('f', s.recv(4))

        name = "" # TODO: implement more efficiently (with a buffer for example)
        while not name.endswith('\n'):
            name += s.recv(1).decode()

        draw_player(x, y, (r, g, b), rad, name)

    pygame.display.update()


def main():
    running = True
    if not handle_menu():
        running = False

    while running:
        response = ""
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                running = False
                pygame.quit()
                return
                
            if event.type == pygame.KEYDOWN:
                if event.key == pygame.K_SPACE:
                    pass # TODO: implement

        keys_pressed = pygame.key.get_pressed()
        if keys_pressed[pygame.K_LEFT]:
            response += 'l'
        if keys_pressed[pygame.K_RIGHT]: 
            response += 'r'
        if keys_pressed[pygame.K_UP]:
            response += 'u'
        if keys_pressed[pygame.K_DOWN]:
            response += 'd'

        recv_and_draw_game()
            
        print(response)
        s.send((response + '\n').encode())
        

if __name__ == "__main__":
    main()


##### TODO LIST #####

# - have some sort of enemy or dangerous obstacle that blows up the ball just like agar
# - space splits ur ball (serverside)
# - player info list should be displayed that describes how big everyone is in order from biggest to smallest. how many players there are and so on.
# - a player should be able to rejoin/respawn, being prompted to enter nickname again, with the old one being prefilled.
# - if a player leaves blow it up and create food that goes in some directions -> food need a velocity
# - also blow up player if it hits obstacle. but maybe dont kill player then
# - hitboxes
# - maybe draw the biggest player last? so that it looks like the biggest player eats the smaller
# - also draw food first. 
# - having an async function that plays background sound MAYBE
# - bots if not enough players has joined. remove these when players join. i.e cap them
# - WINNING CONDITION: when a player gets too big i guess.
# - consider how much the player should grow when it eats another player and when it eats food.
# - ball text should probs change size when ball is bigger or smaller.
# - Work with surfaces in menu to easier make changes to stuff, for example the textfield can be a surface isntead of different items. just to clean up that function cause its ugly af

# You need to regularly make a call to one of four functions in the pygame.event module in order for pygame to internally interact with your OS. Otherwise the OS will think your game has crashed. So make sure you call one of these:

# pygame.event.get() returns a list of all events currently in the event queue.
# pygame.event.poll() returns a single event from the event queue or pygame.NOEVENT if the queue is empty.
# pygame.event.wait() returns a single event from the event queue or waits until an event can be returned.
# pygame.event.pump() allows pygame to handle internal actions. Useful when you don't want to handle events from the event queue.