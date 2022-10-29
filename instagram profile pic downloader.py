import instaloader

ig=instaloader.Instaloader()

dp=input("Enter user name of id : ")

ig.download_profile(dp, profile_pic_only=True)
