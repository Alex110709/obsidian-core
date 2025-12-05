#!/usr/bin/env python3
"""
Obsidian GUI Wallet - Phantom Style

A beautiful graphical wallet for the Obsidian cryptocurrency with Phantom-inspired design.
"""

import tkinter as tk
from tkinter import ttk, messagebox, scrolledtext, filedialog, font
import threading
import time
from wallet import ObsidianWallet


class ModernButton(tk.Canvas):
    """Custom modern button with gradient effect."""
    def __init__(self, parent, text, command, bg_start="#AB47BC", bg_end="#7B1FA2", 
                 fg="#FFFFFF", width=200, height=50, **kwargs):
        super().__init__(parent, width=width, height=height, 
                        highlightthickness=0, **kwargs)
        self.command = command
        self.text = text
        self.bg_start = bg_start
        self.bg_end = bg_end
        self.fg = fg
        
        # Draw gradient button
        self.draw_button()
        
        # Bind events
        self.bind("<Button-1>", self.on_click)
        self.bind("<Enter>", self.on_enter)
        self.bind("<Leave>", self.on_leave)
        self.config(cursor="hand2")
        
    def draw_button(self, hover=False):
        self.delete("all")
        
        # Draw rounded rectangle with gradient effect
        if hover:
            color = "#9C27B0"  # Lighter purple on hover
        else:
            color = self.bg_start
            
        # Rounded rectangle
        self.create_rounded_rect(2, 2, self.winfo_reqwidth()-2, 
                                self.winfo_reqheight()-2, 
                                radius=12, fill=color, outline="")
        
        # Text
        self.create_text(self.winfo_reqwidth()/2, self.winfo_reqheight()/2, 
                        text=self.text, fill=self.fg, 
                        font=("SF Pro Display", 14, "bold"))
    
    def create_rounded_rect(self, x1, y1, x2, y2, radius=25, **kwargs):
        points = [x1+radius, y1,
                 x1+radius, y1,
                 x2-radius, y1,
                 x2-radius, y1,
                 x2, y1,
                 x2, y1+radius,
                 x2, y1+radius,
                 x2, y2-radius,
                 x2, y2-radius,
                 x2, y2,
                 x2-radius, y2,
                 x2-radius, y2,
                 x1+radius, y2,
                 x1+radius, y2,
                 x1, y2,
                 x1, y2-radius,
                 x1, y2-radius,
                 x1, y1+radius,
                 x1, y1+radius,
                 x1, y1]
        return self.create_polygon(points, smooth=True, **kwargs)
    
    def on_click(self, event):
        if self.command:
            self.command()
    
    def on_enter(self, event):
        self.draw_button(hover=True)
    
    def on_leave(self, event):
        self.draw_button(hover=False)


class Card(tk.Frame):
    """Modern card component with shadow effect."""
    def __init__(self, parent, **kwargs):
        super().__init__(parent, bg="#1C1C28", **kwargs)
        self.config(relief=tk.FLAT, bd=0)


class ObsidianGUIWallet:
    def __init__(self, root):
        self.root = root
        self.root.title("Obsidian Wallet")
        self.root.geometry("1100x750")
        
        # Phantom-inspired color scheme
        self.colors = {
            'bg': '#15151F',           # Dark background
            'card': '#1C1C28',          # Card background
            'card_hover': '#25252F',    # Card hover
            'accent': '#AB47BC',        # Purple accent
            'accent_light': '#CE93D8',  # Light purple
            'text': '#FFFFFF',          # White text
            'text_secondary': '#9E9E9E', # Gray text
            'success': '#4CAF50',       # Green
            'warning': '#FF9800',       # Orange
            'error': '#F44336',         # Red
        }
        
        self.root.configure(bg=self.colors['bg'])
        
        # Initialize wallet
        self.wallet = None
        self.wallet_file = "gui_wallet.json"
        
        # Configure styles
        self.setup_styles()
        
        # Create main container
        self.main_container = tk.Frame(root, bg=self.colors['bg'])
        self.main_container.pack(fill=tk.BOTH, expand=True)
        
        # Show initial screen
        self.show_welcome_screen()
        
    def setup_styles(self):
        """Setup custom styles for the GUI."""
        style = ttk.Style()
        style.theme_use('clam')
        
        # Configure colors
        style.configure("TFrame", background=self.colors['bg'])
        style.configure("Card.TFrame", background=self.colors['card'])
        style.configure("TLabel", 
                       background=self.colors['bg'], 
                       foreground=self.colors['text'], 
                       font=("SF Pro Display", 12))
        style.configure("Title.TLabel", 
                       font=("SF Pro Display", 32, "bold"), 
                       foreground=self.colors['text'])
        style.configure("Header.TLabel", 
                       font=("SF Pro Display", 20, "bold"), 
                       foreground=self.colors['text'])
        style.configure("Secondary.TLabel", 
                       foreground=self.colors['text_secondary'], 
                       font=("SF Pro Display", 11))
        
        # Notebook (tabs) styling
        style.configure("TNotebook", 
                       background=self.colors['bg'], 
                       borderwidth=0)
        style.configure("TNotebook.Tab", 
                       background=self.colors['card'],
                       foreground=self.colors['text_secondary'],
                       padding=[20, 10],
                       font=("SF Pro Display", 12))
        style.map("TNotebook.Tab",
                 background=[("selected", self.colors['bg'])],
                 foreground=[("selected", self.colors['accent'])])
        
    def clear_screen(self):
        """Clear the main container."""
        for widget in self.main_container.winfo_children():
            widget.destroy()
            
    def create_card(self, parent, **kwargs):
        """Create a modern card."""
        card = tk.Frame(parent, bg=self.colors['card'], 
                       relief=tk.FLAT, bd=0, **kwargs)
        return card
        
    def show_welcome_screen(self):
        """Show welcome screen with Phantom-style design."""
        self.clear_screen()
        
        # Center container
        center_frame = tk.Frame(self.main_container, bg=self.colors['bg'])
        center_frame.place(relx=0.5, rely=0.5, anchor=tk.CENTER)
        
        # Logo/Icon (using emoji for now)
        logo = tk.Label(center_frame, text="🌑", 
                       font=("SF Pro Display", 80),
                       bg=self.colors['bg'])
        logo.pack(pady=20)
        
        # Title
        title = tk.Label(center_frame, text="Obsidian", 
                        font=("SF Pro Display", 42, "bold"),
                        foreground=self.colors['text'],
                        bg=self.colors['bg'])
        title.pack(pady=10)
        
        # Subtitle
        subtitle = tk.Label(center_frame, 
                           text="A privacy-first cryptocurrency wallet", 
                           font=("SF Pro Display", 14),
                           foreground=self.colors['text_secondary'],
                           bg=self.colors['bg'])
        subtitle.pack(pady=5)
        
        # Buttons container
        buttons_frame = tk.Frame(center_frame, bg=self.colors['bg'])
        buttons_frame.pack(pady=40)
        
        # Create wallet button
        create_btn = ModernButton(buttons_frame, 
                                  text="Create New Wallet",
                                  command=self.show_create_wallet,
                                  bg_start="#AB47BC",
                                  width=280, height=56)
        create_btn.pack(pady=10)
        
        # Load wallet button
        load_btn_frame = tk.Frame(buttons_frame, bg=self.colors['card'], 
                                 highlightthickness=1,
                                 highlightbackground=self.colors['text_secondary'])
        load_btn_frame.pack(pady=10)
        
        load_btn = tk.Label(load_btn_frame, text="Import Existing Wallet",
                           font=("SF Pro Display", 14, "bold"),
                           foreground=self.colors['text'],
                           bg=self.colors['card'],
                           cursor="hand2",
                           padx=40, pady=15)
        load_btn.pack()
        load_btn.bind("<Button-1>", lambda e: self.show_load_wallet())
        
        # RPC settings at bottom
        settings_frame = tk.Frame(self.main_container, bg=self.colors['bg'])
        settings_frame.pack(side=tk.BOTTOM, pady=20)
        
        tk.Label(settings_frame, text="RPC:", 
                foreground=self.colors['text_secondary'],
                bg=self.colors['bg'],
                font=("SF Pro Display", 10)).pack(side=tk.LEFT, padx=5)
        
        self.rpc_host_var = tk.StringVar(value="http://localhost:8545")
        rpc_entry = tk.Entry(settings_frame, textvariable=self.rpc_host_var, 
                            width=30, bg=self.colors['card'], 
                            fg=self.colors['text'],
                            font=("SF Pro Display", 10),
                            relief=tk.FLAT, bd=0)
        rpc_entry.pack(side=tk.LEFT, padx=5, ipady=5)
        
    def show_create_wallet(self):
        """Show screen to create a new wallet."""
        self.clear_screen()
        
        # Header
        header = tk.Frame(self.main_container, bg=self.colors['bg'])
        header.pack(fill=tk.X, padx=40, pady=20)
        
        back_btn = tk.Label(header, text="← Back", 
                           font=("SF Pro Display", 14),
                           foreground=self.colors['accent'],
                           bg=self.colors['bg'],
                           cursor="hand2")
        back_btn.pack(side=tk.LEFT)
        back_btn.bind("<Button-1>", lambda e: self.show_welcome_screen())
        
        # Center content
        content = tk.Frame(self.main_container, bg=self.colors['bg'])
        content.pack(expand=True)
        
        # Icon
        tk.Label(content, text="🔐", 
                font=("SF Pro Display", 60),
                bg=self.colors['bg']).pack(pady=20)
        
        # Title
        tk.Label(content, text="Create New Wallet", 
                font=("SF Pro Display", 28, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack(pady=10)
        
        # Warning card
        warning_card = self.create_card(content)
        warning_card.pack(pady=20, padx=40, fill=tk.X)
        
        tk.Label(warning_card, 
                text="⚠️  Important",
                font=("SF Pro Display", 14, "bold"),
                foreground=self.colors['warning'],
                bg=self.colors['card']).pack(anchor=tk.W, padx=20, pady=(15, 5))
        
        tk.Label(warning_card, 
                text="You will receive a 24-word recovery phrase.\nWrite it down and store it safely - it's the only way to recover your wallet!",
                font=("SF Pro Display", 12),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card'],
                justify=tk.LEFT).pack(anchor=tk.W, padx=20, pady=(0, 15))
        
        # Create button
        create_btn = ModernButton(content, 
                                 text="Generate Wallet",
                                 command=self.create_new_wallet,
                                 width=250, height=50)
        create_btn.pack(pady=30)
        
    def create_new_wallet(self):
        """Create a new wallet and show the mnemonic."""
        try:
            self.wallet = ObsidianWallet(
                rpc_host=self.rpc_host_var.get(),
                wallet_file=self.wallet_file,
                test_mode=False
            )
            
            mnemonic = self.wallet.generate_mnemonic()
            
            # Generate first addresses
            t_addr = self.wallet.generate_transparent_address()
            z_addr = self.wallet.generate_shielded_address()
            
            # Save wallet
            self.wallet.save_wallet(self.wallet_file)
            
            # Show mnemonic
            self.show_mnemonic(mnemonic)
            
        except Exception as e:
            messagebox.showerror("Error", f"Failed to create wallet: {str(e)}")
            
    def show_mnemonic(self, mnemonic):
        """Display the mnemonic phrase in Phantom style."""
        self.clear_screen()
        
        # Header
        header = tk.Frame(self.main_container, bg=self.colors['bg'])
        header.pack(fill=tk.X, padx=40, pady=20)
        
        tk.Label(header, text="Secret Recovery Phrase", 
                font=("SF Pro Display", 24, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack()
        
        # Content
        content = tk.Frame(self.main_container, bg=self.colors['bg'])
        content.pack(expand=True, fill=tk.BOTH, padx=40, pady=20)
        
        # Warning
        warning_card = self.create_card(content)
        warning_card.pack(fill=tk.X, pady=10)
        
        tk.Label(warning_card, 
                text="⚠️  Never share your recovery phrase",
                font=("SF Pro Display", 13, "bold"),
                foreground=self.colors['warning'],
                bg=self.colors['card']).pack(padx=20, pady=15)
        
        # Mnemonic card
        mnemonic_card = self.create_card(content)
        mnemonic_card.pack(fill=tk.BOTH, expand=True, pady=10)
        
        # Grid of words
        words_container = tk.Frame(mnemonic_card, bg=self.colors['card'])
        words_container.pack(padx=30, pady=30, expand=True)
        
        words = mnemonic.split()
        for i, word in enumerate(words):
            row = i // 3
            col = i % 3
            
            word_frame = tk.Frame(words_container, bg=self.colors['bg'], 
                                 relief=tk.FLAT, bd=0)
            word_frame.grid(row=row, column=col, padx=10, pady=8, sticky="ew")
            
            tk.Label(word_frame, 
                    text=f"{i+1}",
                    font=("SF Pro Display", 10),
                    foreground=self.colors['text_secondary'],
                    bg=self.colors['bg'],
                    width=3).pack(side=tk.LEFT, padx=(10, 5))
            
            tk.Label(word_frame, 
                    text=word,
                    font=("SF Pro Mono", 13, "bold"),
                    foreground=self.colors['accent_light'],
                    bg=self.colors['bg']).pack(side=tk.LEFT, padx=5)
        
        # Buttons
        btn_frame = tk.Frame(content, bg=self.colors['bg'])
        btn_frame.pack(fill=tk.X, pady=20)
        
        def copy_mnemonic():
            self.root.clipboard_clear()
            self.root.clipboard_append(mnemonic)
            messagebox.showinfo("✓ Copied", "Recovery phrase copied to clipboard!")
        
        # Copy button
        copy_frame = tk.Frame(btn_frame, bg=self.colors['card'], 
                             highlightthickness=1,
                             highlightbackground=self.colors['accent'])
        copy_frame.pack(side=tk.LEFT, expand=True, fill=tk.X, padx=5)
        
        copy_label = tk.Label(copy_frame, text="📋 Copy to Clipboard",
                             font=("SF Pro Display", 13),
                             foreground=self.colors['text'],
                             bg=self.colors['card'],
                             cursor="hand2",
                             pady=12)
        copy_label.pack()
        copy_label.bind("<Button-1>", lambda e: copy_mnemonic())
        
        # Continue button
        continue_btn = ModernButton(btn_frame, 
                                   text="I've Saved My Phrase",
                                   command=self.show_main_wallet,
                                   width=250, height=50)
        continue_btn.pack(side=tk.RIGHT, padx=5)
        
    def show_load_wallet(self):
        """Show screen to load existing wallet."""
        self.clear_screen()
        
        # Header
        header = tk.Frame(self.main_container, bg=self.colors['bg'])
        header.pack(fill=tk.X, padx=40, pady=20)
        
        back_btn = tk.Label(header, text="← Back", 
                           font=("SF Pro Display", 14),
                           foreground=self.colors['accent'],
                           bg=self.colors['bg'],
                           cursor="hand2")
        back_btn.pack(side=tk.LEFT)
        back_btn.bind("<Button-1>", lambda e: self.show_welcome_screen())
        
        tk.Label(header, text="Import Wallet", 
                font=("SF Pro Display", 24, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack()
        
        # Content
        content = tk.Frame(self.main_container, bg=self.colors['bg'])
        content.pack(expand=True, fill=tk.BOTH, padx=60, pady=20)
        
        # From file card
        file_card = self.create_card(content)
        file_card.pack(fill=tk.X, pady=10)
        
        file_inner = tk.Frame(file_card, bg=self.colors['card'])
        file_inner.pack(padx=30, pady=25)
        
        tk.Label(file_inner, text="Load from File", 
                font=("SF Pro Display", 16, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['card']).pack(pady=10)
        
        tk.Label(file_inner, 
                text="Import a previously saved wallet file",
                font=("SF Pro Display", 12),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(pady=5)
        
        def load_from_file():
            filename = filedialog.askopenfilename(
                title="Select Wallet File",
                filetypes=[("JSON files", "*.json"), ("All files", "*.*")]
            )
            if filename:
                try:
                    self.wallet = ObsidianWallet(
                        rpc_host=self.rpc_host_var.get(),
                        wallet_file=filename,
                        test_mode=False
                    )
                    self.wallet_file = filename
                    self.show_main_wallet()
                except Exception as e:
                    messagebox.showerror("Error", f"Failed to load wallet: {str(e)}")
        
        file_btn = ModernButton(file_inner, 
                               text="Browse Files",
                               command=load_from_file,
                               bg_start="#5E35B1",
                               width=200, height=45)
        file_btn.pack(pady=15)
        
        # Divider
        tk.Label(content, text="OR", 
                font=("SF Pro Display", 12),
                foreground=self.colors['text_secondary'],
                bg=self.colors['bg']).pack(pady=15)
        
        # From mnemonic card
        mnemonic_card = self.create_card(content)
        mnemonic_card.pack(fill=tk.BOTH, expand=True, pady=10)
        
        mnemonic_inner = tk.Frame(mnemonic_card, bg=self.colors['card'])
        mnemonic_inner.pack(padx=30, pady=25, fill=tk.BOTH, expand=True)
        
        tk.Label(mnemonic_inner, text="Restore from Recovery Phrase", 
                font=("SF Pro Display", 16, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['card']).pack(pady=10)
        
        tk.Label(mnemonic_inner, 
                text="Enter your 24-word recovery phrase",
                font=("SF Pro Display", 12),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(pady=5)
        
        # Text area
        text_frame = tk.Frame(mnemonic_inner, bg=self.colors['bg'])
        text_frame.pack(pady=15, fill=tk.BOTH, expand=True)
        
        mnemonic_text = scrolledtext.ScrolledText(text_frame, height=4, width=60,
                                                  bg=self.colors['bg'], 
                                                  fg=self.colors['text'],
                                                  font=("SF Pro Mono", 12),
                                                  relief=tk.FLAT,
                                                  insertbackground=self.colors['accent'])
        mnemonic_text.pack(fill=tk.BOTH, expand=True, padx=10, pady=10)
        
        def restore_from_mnemonic():
            mnemonic = mnemonic_text.get("1.0", tk.END).strip()
            if not mnemonic:
                messagebox.showwarning("Warning", "Please enter your recovery phrase")
                return
                
            try:
                self.wallet = ObsidianWallet(
                    rpc_host=self.rpc_host_var.get(),
                    wallet_file=self.wallet_file,
                    test_mode=False
                )
                
                if self.wallet.load_mnemonic(mnemonic):
                    self.wallet.generate_transparent_address()
                    self.wallet.generate_shielded_address()
                    self.wallet.save_wallet(self.wallet_file)
                    self.show_main_wallet()
                else:
                    messagebox.showerror("Error", "Invalid recovery phrase")
            except Exception as e:
                messagebox.showerror("Error", f"Failed to restore wallet: {str(e)}")
        
        restore_btn = ModernButton(mnemonic_inner, 
                                  text="Restore Wallet",
                                  command=restore_from_mnemonic,
                                  width=200, height=45)
        restore_btn.pack(pady=10)
        
    def show_main_wallet(self):
        """Show main wallet interface with Phantom-style tabs."""
        self.clear_screen()
        
        # Top bar
        topbar = tk.Frame(self.main_container, bg=self.colors['bg'], height=70)
        topbar.pack(fill=tk.X, padx=20, pady=(10, 0))
        topbar.pack_propagate(False)
        
        # Logo
        tk.Label(topbar, text="🌑 Obsidian", 
                font=("SF Pro Display", 18, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack(side=tk.LEFT, padx=10)
        
        # Create notebook for tabs
        notebook = ttk.Notebook(self.main_container)
        notebook.pack(fill=tk.BOTH, expand=True, padx=20, pady=10)
        
        # Portfolio Tab
        portfolio_frame = tk.Frame(notebook, bg=self.colors['bg'])
        notebook.add(portfolio_frame, text="  Portfolio  ")
        self.create_portfolio_tab(portfolio_frame)
        
        # Send Tab
        send_frame = tk.Frame(notebook, bg=self.colors['bg'])
        notebook.add(send_frame, text="  Send  ")
        self.create_send_tab(send_frame)
        
        # Receive Tab
        receive_frame = tk.Frame(notebook, bg=self.colors['bg'])
        notebook.add(receive_frame, text="  Receive  ")
        self.create_receive_tab(receive_frame)
        
        # Activity Tab
        activity_frame = tk.Frame(notebook, bg=self.colors['bg'])
        notebook.add(activity_frame, text="  Activity  ")
        self.create_history_tab(activity_frame)
        
        # Settings Tab
        settings_frame = tk.Frame(notebook, bg=self.colors['bg'])
        notebook.add(settings_frame, text="  Settings  ")
        self.create_settings_tab(settings_frame)
        
    def create_portfolio_tab(self, parent):
        """Create the portfolio tab with Phantom style."""
        # Scrollable container
        canvas = tk.Canvas(parent, bg=self.colors['bg'], highlightthickness=0)
        scrollbar = tk.Scrollbar(parent, orient="vertical", command=canvas.yview)
        scrollable_frame = tk.Frame(canvas, bg=self.colors['bg'])
        
        scrollable_frame.bind(
            "<Configure>",
            lambda e: canvas.configure(scrollregion=canvas.bbox("all"))
        )
        
        canvas.create_window((0, 0), window=scrollable_frame, anchor="nw")
        canvas.configure(yscrollcommand=scrollbar.set)
        
        # Balance card
        balance_card = self.create_card(scrollable_frame)
        balance_card.pack(fill=tk.X, padx=20, pady=20)
        
        balance_inner = tk.Frame(balance_card, bg=self.colors['card'])
        balance_inner.pack(padx=30, pady=30)
        
        tk.Label(balance_inner, text="Total Balance", 
                font=("SF Pro Display", 14),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack()
        
        self.balance_label = tk.Label(balance_inner, text="0.00000000", 
                                      font=("SF Pro Display", 48, "bold"),
                                      foreground=self.colors['text'],
                                      bg=self.colors['card'])
        self.balance_label.pack(pady=10)
        
        tk.Label(balance_inner, text="OBS", 
                font=("SF Pro Display", 18),
                foreground=self.colors['accent'],
                bg=self.colors['card']).pack()
        
        # Refresh button
        refresh_frame = tk.Frame(balance_inner, bg=self.colors['bg'])
        refresh_frame.pack(pady=15)
        
        refresh_btn = tk.Label(refresh_frame, text="🔄 Refresh",
                              font=("SF Pro Display", 12),
                              foreground=self.colors['accent'],
                              bg=self.colors['bg'],
                              cursor="hand2",
                              padx=15, pady=8)
        refresh_btn.pack()
        refresh_btn.bind("<Button-1>", lambda e: self.refresh_balances())
        
        # Addresses section
        tk.Label(scrollable_frame, text="Your Addresses", 
                font=("SF Pro Display", 18, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack(anchor=tk.W, padx=20, pady=(20, 10))
        
        # Addresses list frame
        self.addresses_list_frame = tk.Frame(scrollable_frame, bg=self.colors['bg'])
        self.addresses_list_frame.pack(fill=tk.BOTH, expand=True, padx=20, pady=10)
        
        canvas.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        
        # Initial load
        self.refresh_balances()
        
    def refresh_balances(self):
        """Refresh all balances with modern card design."""
        def update():
            try:
                # Clear existing address cards
                for widget in self.addresses_list_frame.winfo_children():
                    widget.destroy()
                
                total_balance = 0.0
                
                # Get transparent addresses
                for priv, pub, addr in self.wallet.transparent_keys:
                    try:
                        balance_info = self.wallet.get_balance(addr)
                        balance = balance_info if isinstance(balance_info, float) else 0.0
                        total_balance += balance
                        
                        # Create address card
                        self.create_address_card(self.addresses_list_frame, 
                                                "Transparent", addr, balance)
                    except:
                        self.create_address_card(self.addresses_list_frame, 
                                                "Transparent", addr, 0.0, error=True)
                
                # Get shielded addresses
                for pub, view, addr in self.wallet.shielded_keys:
                    try:
                        balance_info = self.wallet.get_balance(addr)
                        balance = balance_info if isinstance(balance_info, float) else 0.0
                        total_balance += balance
                        
                        self.create_address_card(self.addresses_list_frame, 
                                                "Shielded", addr, balance, private=True)
                    except:
                        self.create_address_card(self.addresses_list_frame, 
                                                "Shielded", addr, 0.0, private=True, error=True)
                
                # Update total balance
                self.balance_label.config(text=f"{total_balance:.8f}")
                
            except Exception as e:
                print(f"Error refreshing balances: {e}")
        
        # Run in background thread
        threading.Thread(target=update, daemon=True).start()
    
    def create_address_card(self, parent, addr_type, address, balance, private=False, error=False):
        """Create a modern address card."""
        card = self.create_card(parent)
        card.pack(fill=tk.X, pady=5)
        
        inner = tk.Frame(card, bg=self.colors['card'])
        inner.pack(fill=tk.X, padx=20, pady=15)
        
        # Left side - icon and type
        left = tk.Frame(inner, bg=self.colors['card'])
        left.pack(side=tk.LEFT)
        
        icon = "🔒" if private else "👁️"
        tk.Label(left, text=icon, 
                font=("SF Pro Display", 24),
                bg=self.colors['card']).pack(side=tk.LEFT, padx=(0, 10))
        
        info = tk.Frame(left, bg=self.colors['card'])
        info.pack(side=tk.LEFT)
        
        tk.Label(info, text=addr_type, 
                font=("SF Pro Display", 13, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['card']).pack(anchor=tk.W)
        
        addr_label = tk.Label(info, 
                             text=address[:20] + "..." + address[-10:],
                             font=("SF Pro Mono", 10),
                             foreground=self.colors['text_secondary'],
                             bg=self.colors['card'])
        addr_label.pack(anchor=tk.W)
        
        # Right side - balance
        if error:
            tk.Label(inner, text="Error", 
                    font=("SF Pro Display", 14),
                    foreground=self.colors['error'],
                    bg=self.colors['card']).pack(side=tk.RIGHT)
        else:
            tk.Label(inner, text=f"{balance:.8f} OBS", 
                    font=("SF Pro Display", 14, "bold"),
                    foreground=self.colors['text'],
                    bg=self.colors['card']).pack(side=tk.RIGHT)
        
    def create_send_tab(self, parent):
        """Create the send tab with modern design."""
        # Center container
        center = tk.Frame(parent, bg=self.colors['bg'])
        center.place(relx=0.5, rely=0.5, anchor=tk.CENTER)
        
        # Title
        tk.Label(center, text="Send OBS", 
                font=("SF Pro Display", 24, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack(pady=20)
        
        # Form card
        form_card = self.create_card(center)
        form_card.pack(pady=10)
        
        form_inner = tk.Frame(form_card, bg=self.colors['card'])
        form_inner.pack(padx=40, pady=30)
        
        # From
        tk.Label(form_inner, text="From", 
                font=("SF Pro Display", 12, "bold"),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(anchor=tk.W, pady=(0, 5))
        
        self.send_from_var = tk.StringVar()
        from_combo = ttk.Combobox(form_inner, textvariable=self.send_from_var, 
                                 width=50, font=("SF Pro Display", 12))
        from_combo['values'] = self.get_all_addresses()
        if from_combo['values']:
            from_combo.current(0)
        from_combo.pack(pady=(0, 20), ipady=8)
        
        # To
        tk.Label(form_inner, text="To", 
                font=("SF Pro Display", 12, "bold"),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(anchor=tk.W, pady=(0, 5))
        
        self.send_to_var = tk.StringVar()
        to_entry = tk.Entry(form_inner, textvariable=self.send_to_var, width=53,
                           bg=self.colors['bg'], fg=self.colors['text'],
                           font=("SF Pro Mono", 12),
                           relief=tk.FLAT, bd=0, insertbackground=self.colors['accent'])
        to_entry.pack(pady=(0, 20), ipady=10)
        
        # Amount
        tk.Label(form_inner, text="Amount", 
                font=("SF Pro Display", 12, "bold"),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(anchor=tk.W, pady=(0, 5))
        
        amount_frame = tk.Frame(form_inner, bg=self.colors['bg'])
        amount_frame.pack(fill=tk.X, pady=(0, 20))
        
        self.send_amount_var = tk.StringVar()
        amount_entry = tk.Entry(amount_frame, textvariable=self.send_amount_var,
                               bg=self.colors['bg'], fg=self.colors['text'],
                               font=("SF Pro Display", 16),
                               relief=tk.FLAT, bd=0, insertbackground=self.colors['accent'])
        amount_entry.pack(side=tk.LEFT, fill=tk.X, expand=True, ipady=10, padx=(10, 5))
        
        tk.Label(amount_frame, text="OBS",
                font=("SF Pro Display", 14),
                foreground=self.colors['text_secondary'],
                bg=self.colors['bg']).pack(side=tk.RIGHT, padx=10)
        
        # Memo
        tk.Label(form_inner, text="Memo (optional)", 
                font=("SF Pro Display", 12, "bold"),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(anchor=tk.W, pady=(0, 5))
        
        self.send_memo_var = tk.StringVar()
        memo_entry = tk.Entry(form_inner, textvariable=self.send_memo_var, width=53,
                             bg=self.colors['bg'], fg=self.colors['text'],
                             font=("SF Pro Display", 12),
                             relief=tk.FLAT, bd=0, insertbackground=self.colors['accent'])
        memo_entry.pack(pady=(0, 30), ipady=10)
        
        # Send button
        def send_transaction():
            from_addr = self.send_from_var.get()
            to_addr = self.send_to_var.get()
            amount_str = self.send_amount_var.get()
            memo = self.send_memo_var.get()
            
            if not from_addr or not to_addr or not amount_str:
                messagebox.showwarning("Warning", "Please fill in all required fields")
                return
            
            try:
                amount = float(amount_str)
                if amount <= 0:
                    messagebox.showwarning("Warning", "Amount must be positive")
                    return
                
                if messagebox.askyesno("Confirm", 
                                      f"Send {amount} OBS to\n{to_addr[:30]}...?"):
                    txid = self.wallet.send_transaction(from_addr, to_addr, amount, memo)
                    messagebox.showinfo("✓ Success", f"Transaction sent!\nTXID: {txid[:16]}...")
                    
                    self.send_to_var.set("")
                    self.send_amount_var.set("")
                    self.send_memo_var.set("")
                    self.refresh_balances()
                
            except Exception as e:
                messagebox.showerror("Error", f"Failed to send: {str(e)}")
        
        send_btn = ModernButton(form_inner, 
                               text="Send Transaction",
                               command=send_transaction,
                               width=300, height=50)
        send_btn.pack()
        
    def create_receive_tab(self, parent):
        """Create the receive tab."""
        center = tk.Frame(parent, bg=self.colors['bg'])
        center.place(relx=0.5, rely=0.5, anchor=tk.CENTER)
        
        tk.Label(center, text="Receive OBS", 
                font=("SF Pro Display", 24, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack(pady=20)
        
        # Type selection card
        type_card = self.create_card(center)
        type_card.pack(pady=10, ipadx=30, ipady=20)
        
        self.receive_type_var = tk.StringVar(value="transparent")
        
        tk.Label(type_card, text="Select Address Type", 
                font=("SF Pro Display", 14, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['card']).pack(pady=(10, 15))
        
        radio_frame = tk.Frame(type_card, bg=self.colors['card'])
        radio_frame.pack()
        
        tk.Radiobutton(radio_frame, text="👁️ Transparent (Public)", 
                      variable=self.receive_type_var, value="transparent",
                      bg=self.colors['card'], fg=self.colors['text'],
                      selectcolor=self.colors['bg'],
                      activebackground=self.colors['card'],
                      activeforeground=self.colors['text'],
                      font=("SF Pro Display", 13)).pack(anchor=tk.W, pady=5, padx=20)
        
        tk.Radiobutton(radio_frame, text="🔒 Shielded (Private)", 
                      variable=self.receive_type_var, value="shielded",
                      bg=self.colors['card'], fg=self.colors['text'],
                      selectcolor=self.colors['bg'],
                      activebackground=self.colors['card'],
                      activeforeground=self.colors['text'],
                      font=("SF Pro Display", 13)).pack(anchor=tk.W, pady=5, padx=20)
        
        # Generate button
        def generate_address():
            addr_type = self.receive_type_var.get()
            try:
                if addr_type == "transparent":
                    addr = self.wallet.generate_transparent_address()
                else:
                    addr = self.wallet.generate_shielded_address()
                
                self.wallet.save_wallet(self.wallet_file)
                self.receive_addr_var.set(addr)
                messagebox.showinfo("✓ Success", f"New {addr_type} address generated!")
                self.refresh_balances()
                
            except Exception as e:
                messagebox.showerror("Error", f"Failed: {str(e)}")
        
        gen_btn = ModernButton(center, 
                              text="Generate New Address",
                              command=generate_address,
                              width=250, height=50)
        gen_btn.pack(pady=20)
        
        # Address display card
        addr_card = self.create_card(center)
        addr_card.pack(pady=10, fill=tk.X, padx=40)
        
        addr_inner = tk.Frame(addr_card, bg=self.colors['card'])
        addr_inner.pack(padx=30, pady=25)
        
        tk.Label(addr_inner, text="Your Address", 
                font=("SF Pro Display", 12),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(pady=(0, 10))
        
        self.receive_addr_var = tk.StringVar()
        addr_display = tk.Entry(addr_inner, textvariable=self.receive_addr_var,
                               width=50, bg=self.colors['bg'],
                               fg=self.colors['accent_light'],
                               font=("SF Pro Mono", 12, "bold"),
                               relief=tk.FLAT, bd=0,
                               state="readonly", justify=tk.CENTER)
        addr_display.pack(ipady=10)
        
        # Copy button
        def copy_address():
            addr = self.receive_addr_var.get()
            if addr:
                self.root.clipboard_clear()
                self.root.clipboard_append(addr)
                messagebox.showinfo("✓ Copied", "Address copied!")
        
        copy_frame = tk.Frame(addr_inner, bg=self.colors['bg'])
        copy_frame.pack(pady=15)
        
        copy_btn = tk.Label(copy_frame, text="📋 Copy",
                           font=("SF Pro Display", 12),
                           foreground=self.colors['accent'],
                           bg=self.colors['bg'],
                           cursor="hand2",
                           padx=20, pady=8)
        copy_btn.pack()
        copy_btn.bind("<Button-1>", lambda e: copy_address())
        
    def create_history_tab(self, parent):
        """Create the activity/history tab."""
        # Header
        header = tk.Frame(parent, bg=self.colors['bg'])
        header.pack(fill=tk.X, padx=30, pady=20)
        
        tk.Label(header, text="Activity", 
                font=("SF Pro Display", 22, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack(side=tk.LEFT)
        
        # Refresh button
        refresh_btn = tk.Label(header, text="🔄",
                              font=("SF Pro Display", 18),
                              foreground=self.colors['accent'],
                              bg=self.colors['bg'],
                              cursor="hand2")
        refresh_btn.pack(side=tk.RIGHT)
        refresh_btn.bind("<Button-1>", lambda e: self.refresh_history())
        
        # Scrollable list
        canvas = tk.Canvas(parent, bg=self.colors['bg'], highlightthickness=0)
        scrollbar = tk.Scrollbar(parent, orient="vertical", command=canvas.yview)
        
        self.history_frame = tk.Frame(canvas, bg=self.colors['bg'])
        
        self.history_frame.bind(
            "<Configure>",
            lambda e: canvas.configure(scrollregion=canvas.bbox("all"))
        )
        
        canvas.create_window((0, 0), window=self.history_frame, anchor="nw")
        canvas.configure(yscrollcommand=scrollbar.set)
        
        canvas.pack(side=tk.LEFT, fill=tk.BOTH, expand=True, padx=20)
        scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        
        self.refresh_history()
    
    def refresh_history(self):
        """Refresh transaction history."""
        for widget in self.history_frame.winfo_children():
            widget.destroy()
        
        if not self.wallet.transaction_history:
            tk.Label(self.history_frame, 
                    text="No transactions yet",
                    font=("SF Pro Display", 14),
                    foreground=self.colors['text_secondary'],
                    bg=self.colors['bg']).pack(pady=50)
            return
        
        for tx in reversed(self.wallet.transaction_history):
            self.create_transaction_card(self.history_frame, tx)
    
    def create_transaction_card(self, parent, tx):
        """Create a transaction card."""
        card = self.create_card(parent)
        card.pack(fill=tk.X, pady=5)
        
        inner = tk.Frame(card, bg=self.colors['card'])
        inner.pack(fill=tk.X, padx=20, pady=15)
        
        # Left - icon and info
        left = tk.Frame(inner, bg=self.colors['card'])
        left.pack(side=tk.LEFT, fill=tk.X, expand=True)
        
        # Icon based on type
        icon = "📤" if tx.get('from') else "📥"
        tk.Label(left, text=icon, 
                font=("SF Pro Display", 20),
                bg=self.colors['card']).pack(side=tk.LEFT, padx=(0, 10))
        
        info = tk.Frame(left, bg=self.colors['card'])
        info.pack(side=tk.LEFT, fill=tk.X, expand=True)
        
        # Type and status
        tx_type = tx.get('type', 'Transaction')
        tk.Label(info, text=tx_type.capitalize(), 
                font=("SF Pro Display", 13, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['card']).pack(anchor=tk.W)
        
        # Time
        timestamp = time.strftime('%b %d, %Y %H:%M', 
                                 time.localtime(tx['timestamp']))
        tk.Label(info, text=timestamp,
                font=("SF Pro Display", 11),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(anchor=tk.W)
        
        # Right - amount
        amount_color = self.colors['success'] if icon == "📥" else self.colors['text']
        amount_prefix = "+" if icon == "📥" else "-"
        
        tk.Label(inner, 
                text=f"{amount_prefix}{tx['amount']:.4f} OBS",
                font=("SF Pro Display", 14, "bold"),
                foreground=amount_color,
                bg=self.colors['card']).pack(side=tk.RIGHT)
        
    def create_settings_tab(self, parent):
        """Create the settings tab."""
        # Scrollable
        canvas = tk.Canvas(parent, bg=self.colors['bg'], highlightthickness=0)
        scrollbar = tk.Scrollbar(parent, orient="vertical", command=canvas.yview)
        scrollable = tk.Frame(canvas, bg=self.colors['bg'])
        
        scrollable.bind(
            "<Configure>",
            lambda e: canvas.configure(scrollregion=canvas.bbox("all"))
        )
        
        canvas.create_window((0, 0), window=scrollable, anchor="nw")
        canvas.configure(yscrollcommand=scrollbar.set)
        
        # Title
        tk.Label(scrollable, text="Settings", 
                font=("SF Pro Display", 22, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['bg']).pack(anchor=tk.W, padx=30, pady=20)
        
        # RPC settings card
        rpc_card = self.create_card(scrollable)
        rpc_card.pack(fill=tk.X, padx=30, pady=10)
        
        rpc_inner = tk.Frame(rpc_card, bg=self.colors['card'])
        rpc_inner.pack(padx=25, pady=20)
        
        tk.Label(rpc_inner, text="Network", 
                font=("SF Pro Display", 16, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['card']).pack(anchor=tk.W, pady=(0, 15))
        
        tk.Label(rpc_inner, text="RPC Host", 
                font=("SF Pro Display", 12),
                foreground=self.colors['text_secondary'],
                bg=self.colors['card']).pack(anchor=tk.W, pady=(0, 5))
        
        rpc_entry = tk.Entry(rpc_inner, textvariable=self.rpc_host_var,
                            width=50, bg=self.colors['bg'],
                            fg=self.colors['text'],
                            font=("SF Pro Mono", 11),
                            relief=tk.FLAT, bd=0,
                            insertbackground=self.colors['accent'])
        rpc_entry.pack(anchor=tk.W, ipady=8)
        
        # Wallet info card
        info_card = self.create_card(scrollable)
        info_card.pack(fill=tk.X, padx=30, pady=10)
        
        info_inner = tk.Frame(info_card, bg=self.colors['card'])
        info_inner.pack(padx=25, pady=20)
        
        tk.Label(info_inner, text="Wallet Info", 
                font=("SF Pro Display", 16, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['card']).pack(anchor=tk.W, pady=(0, 15))
        
        info_items = [
            ("File", self.wallet_file),
            ("Transparent Addresses", str(len(self.wallet.transparent_keys))),
            ("Shielded Addresses", str(len(self.wallet.shielded_keys)))
        ]
        
        for label, value in info_items:
            row = tk.Frame(info_inner, bg=self.colors['card'])
            row.pack(fill=tk.X, pady=5)
            
            tk.Label(row, text=label,
                    font=("SF Pro Display", 12),
                    foreground=self.colors['text_secondary'],
                    bg=self.colors['card']).pack(side=tk.LEFT)
            
            tk.Label(row, text=value,
                    font=("SF Pro Display", 12),
                    foreground=self.colors['text'],
                    bg=self.colors['card']).pack(side=tk.RIGHT)
        
        # Actions
        actions_card = self.create_card(scrollable)
        actions_card.pack(fill=tk.X, padx=30, pady=10)
        
        actions_inner = tk.Frame(actions_card, bg=self.colors['card'])
        actions_inner.pack(padx=25, pady=20)
        
        tk.Label(actions_inner, text="Actions", 
                font=("SF Pro Display", 16, "bold"),
                foreground=self.colors['text'],
                bg=self.colors['card']).pack(anchor=tk.W, pady=(0, 15))
        
        # Backup button
        def backup_wallet():
            filename = filedialog.asksaveasfilename(
                defaultextension=".json",
                filetypes=[("JSON files", "*.json")],
                initialfile="obsidian_backup.json"
            )
            if filename:
                try:
                    self.wallet.save_wallet(filename)
                    messagebox.showinfo("✓ Success", "Wallet backed up!")
                except Exception as e:
                    messagebox.showerror("Error", f"Failed: {str(e)}")
        
        backup_btn = ModernButton(actions_inner, 
                                 text="💾 Backup Wallet",
                                 command=backup_wallet,
                                 bg_start="#5E35B1",
                                 width=250, height=45)
        backup_btn.pack(pady=5)
        
        # Show recovery phrase
        def show_recovery():
            if messagebox.askyesno("⚠️ Warning", 
                                  "Never share your recovery phrase!\n\nContinue?"):
                # Create popup
                popup = tk.Toplevel(self.root)
                popup.title("Recovery Phrase")
                popup.geometry("600x400")
                popup.configure(bg=self.colors['bg'])
                
                tk.Label(popup, text="Secret Recovery Phrase",
                        font=("SF Pro Display", 18, "bold"),
                        foreground=self.colors['text'],
                        bg=self.colors['bg']).pack(pady=20)
                
                text = tk.Text(popup, height=10, width=50,
                              bg=self.colors['card'],
                              fg=self.colors['accent_light'],
                              font=("SF Pro Mono", 12),
                              relief=tk.FLAT, bd=0)
                text.pack(pady=20, padx=30)
                text.insert("1.0", self.wallet.mnemonic)
                text.config(state=tk.DISABLED)
                
                ModernButton(popup, text="Close", 
                           command=popup.destroy,
                           width=150, height=40).pack(pady=20)
        
        recovery_btn = ModernButton(actions_inner, 
                                   text="🔑 Show Recovery Phrase",
                                   command=show_recovery,
                                   bg_start="#D32F2F",
                                   width=250, height=45)
        recovery_btn.pack(pady=5)
        
        canvas.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        
    def get_all_addresses(self):
        """Get all wallet addresses."""
        addresses = []
        for _, _, addr in self.wallet.transparent_keys:
            addresses.append(addr)
        for _, _, addr in self.wallet.shielded_keys:
            addresses.append(addr)
        return addresses


def main():
    root = tk.Tk()
    
    # Set app icon (if available)
    try:
        # You can add a .ico file here
        # root.iconbitmap('icon.ico')
        pass
    except:
        pass
    
    app = ObsidianGUIWallet(root)
    root.mainloop()


if __name__ == "__main__":
    main()
