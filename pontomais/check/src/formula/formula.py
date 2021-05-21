#!/usr/bin/python3

import time
from selenium import webdriver
from selenium.webdriver.chrome.options import Options

def Run(email, password):
    chrome_options = Options()
    chrome_options.add_argument('--headless')
    chrome_options.add_argument('--no-sandbox')
    chrome_options.add_argument('--disable-dev-shm-usage')

    driver = webdriver.Chrome(chrome_options=chrome_options, executable_path="./chromedriver")

    driver.get('https://app.pontomaisweb.com.br/#/acessar')

    emailElement = driver.find_element_by_name("login")
    emailElement.send_keys(email)

    passwordElement = driver.find_element_by_name("password")
    passwordElement.send_keys(password)

    submitLoginElement = driver.find_element_by_class_name("btn-primary")
    submitLoginElement.click()

    driver.implicitly_wait(100)

    submitMarkElement = driver.find_element_by_xpath('//button[text()="Registrar ponto"]')
    submitMarkElement.click()

    time.sleep(2)

    driver.close()